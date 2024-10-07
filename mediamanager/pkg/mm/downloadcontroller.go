package mm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/rpc"
	"time"

	"github.com/cenkalti/rain/rainrpc"
)

type DownloadController struct {
	Torrents  *rainrpc.Client
	Downloads DownloadStore
	Logger    *slog.Logger
	Trackers  []string
}

func (c *DownloadController) Run(
	ctx context.Context,
	interval time.Duration,
) error {
	c.Logger.Info("starting download controller")

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		if err := c.RunLoop(ctx); err != nil {
			if errors.Is(err, context.Canceled) ||
				errors.Is(err, context.DeadlineExceeded) {
				return err
			}
			c.Logger.Error("running control loop", "err", err.Error())
		}
	}

	return nil
}

func (c *DownloadController) RunLoop(ctx context.Context) error {
	incomplete, err := c.listIncompleteDownloads(ctx)
	if err != nil {
		return err
	}

	torrents, err := c.Torrents.ListTorrents()
	if err != nil {
		return err
	}

	seen := make(map[InfoHash]struct{})
	// this is going to iterate over the list of torrents which may
	// contain duplicates for a given infohash because `rain` allows
	// duplicates. we will still fetch the filestats for all duplicates
	// and then downstream of that, we will populate the download with
	// information from the most complete torrent for that infohash.
	var update []torrent
	for i := range torrents {
		infoHash := NewInfoHash(torrents[i].InfoHash)
		seen[infoHash] = struct{}{}
		if d, exists := incomplete[infoHash]; exists {
			update = append(
				update,
				torrent{id: torrents[i].ID, download: d},
			)
		}
	}

	results := make(chan error)
	go func() { results <- c.updateDownloads(ctx, update) }()
	go func() { results <- c.createTorrents(ctx, incomplete, seen) }()
	for range 2 {
		// the goroutines only return context errors, so we want to return
		// if we get an error.
		if err := <-results; err != nil {
			return fmt.Errorf("syncing downloads: %w", err)
		}
	}

	return nil
}

func (c *DownloadController) listIncompleteDownloads(
	ctx context.Context,
) (map[InfoHash]*Download, error) {
	downloads, err := c.Downloads.ListDownloads(ctx)
	if err != nil {
		return nil, err
	}

	incomplete := make(map[InfoHash]*Download)
	for i := range downloads {
		if downloads[i].Status != DownloadStatusComplete {
			incomplete[downloads[i].ID] = &downloads[i]
		}
	}
	return incomplete, nil
}

func (c *DownloadController) createTorrents(
	ctx context.Context,
	filtered map[InfoHash]*Download,
	seen map[InfoHash]struct{},
) error {
	// check to see if there are any in-progress downloads which don't
	// have corresponding torrents and create them as necessary
	for _, download := range filtered {
		if _, exists := seen[download.ID]; !exists {
			if err := ctx.Err(); err != nil {
				return err
			}
			if _, err := c.Torrents.AddURI(
				Magnet("", download.ID, c.Trackers...),
				&rainrpc.AddTorrentOptions{ID: download.ID.String()},
			); err != nil {
				c.Logger.Error(
					"creating torrent for download",
					"err", err.Error(),
					"infoHash", download.ID,
				)
				continue
			}
		}
	}
	return nil
}

func (c *DownloadController) updateDownloads(
	ctx context.Context,
	torrents []torrent,
) error {
	downloads := make(map[InfoHash]*Download)
	for _, torrent := range torrents {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf(
				"fetching file stats for torrent `%s`: %w",
				torrent.id,
				err,
			)
		}

		fileStats, err := c.Torrents.GetTorrentFileStats(torrent.id)
		if err != nil {
			pending, err := isAwaitingMetadata(err)
			if err != nil {
				c.Logger.Error(
					"fetching file stats for torrent",
					"err", err.Error(),
					"torrent", torrent.id,
					"infoHash", torrent.download.ID,
				)
				continue
			}
			if pending {
				// we know the torrent is awaiting metadata, but we can't
				// automatically update the status of the download at this
				// point because there may be another torrent for the same
				// download which is farther along.
				continue
			}
		}

		prospective := Download{
			ID:     torrent.download.ID,
			Status: DownloadStatusProgress,
			Files:  make(DownloadFiles, len(fileStats)),
		}
		for i := range fileStats {
			prospective.Files[i] = DownloadFile{
				Path:     fileStats[i].File.Path,
				Progress: uint64(fileStats[i].BytesCompleted),
				Size:     uint64(fileStats[i].File.Length),
			}
			prospective.Size += uint64(fileStats[i].File.Length)
			prospective.Progress += uint64(fileStats[i].BytesCompleted)
		}

		if prospective.Progress >= prospective.Size {
			prospective.Status = DownloadStatusComplete
		}

		_, exists := downloads[torrent.download.ID]
		slog.Debug(
			"deciding whether or not to update download",
			"exists", exists,
			"prospectiveProgress", prospective.Progress,
			"prospectiveSize", prospective.Size,
			"torrentDownloadProgress", torrent.download.Progress,
		)
		if !exists || prospective.Progress >= torrent.download.Progress {
			downloads[torrent.download.ID] = &prospective
		}
	}

	list := make([]Download, 0, len(downloads))
	for _, d := range downloads {
		list = append(list, *d)
	}

	if err := c.Downloads.PutDownloads(ctx, list); err != nil {
		c.Logger.Error(
			"updating downloads based on torrents",
			"err", err.Error(),
		)
	}

	return nil
}

func isAwaitingMetadata(err error) (bool, error) {
	// `rainrpc` returns an `rpc.ServerError` which is a string, but the string
	// itself is the JSON representation of a `*jsonrpc2.Error` which in turn
	// has a `.message` property containing the result of calling the `Error()`
	// method on the actual rain error.
	// * [the upstream rain error](https://github.com/cenkalti/rain/blob/22bbeee110d60a9077a6a5239807ee39156dd059/torrent/torrent.go#L423)
	// * [`jsonrpc2.Error` type](https://github.com/powerman/rpc-codec/blob/master/jsonrpc2/errors.go#L23-L27)
	var serverErr *rpc.ServerError
	if errors.As(err, &serverErr) {
		var payload struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal(
			[]byte(serverErr.Error()),
			&payload,
		); err != nil {
			return false, fmt.Errorf("assessing metadata status: %w", err)
		}

		return payload.Message == "torrent metadata not ready", nil
	}

	return false, nil
}

type torrent struct {
	id       string
	download *Download
}
