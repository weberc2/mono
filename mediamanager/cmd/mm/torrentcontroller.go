package main

import (
	"context"
	"fmt"

	"github.com/cenkalti/rain/rainrpc"
)

type TorrentController struct {
	Torrents  *rainrpc.Client
	Downloads DownloadStore
	Register  interface {
		RegisteredDownloads(ctx Context) (map[DownloadID]struct{}, error)
	}
}

func (c *TorrentController) RunLoop(ctx context.Context) error {
	type torrent struct {
		id       string
		name     string
		infoHash InfoHash
		files    []DownloadFile
		size     uint64
		progress uint64
	}

	var torrents []torrent
	var downloads []Download
	downloadsByInfoHash := make(chan map[InfoHash]*Download)
	fileStatsRequests := make(chan *torrent, 1)
	updatedTorrents := make(chan *torrent, 1)
	parallel(
		ctx,
		func(ctx Context) error {
			defer close(downloadsByInfoHash)
			var err error
			if downloads, err = c.Downloads.ListDownloads(ctx); err != nil {
				return err
			}
			byInfoHash := make(map[InfoHash]*Download)
			for i := range downloads {
				byInfoHash[downloads[i].InfoHash] = &downloads[i]
			}
			<-downloadsByInfoHash
			return nil
		},
		func(ctx Context) error {
			defer close(fileStatsRequests)
			ts, err := c.Torrents.ListTorrents()
			if err != nil {
				return err
			}

			// should execute 0 or 1 times (0 if there was an error fetching
			// downloads list)
			for byInfoHash := range downloadsByInfoHash {
				torrents = make([]torrent, 0, len(ts))
				for i := range ts {
					if dl, exists := byInfoHash[InfoHash(
						ts[i].InfoHash,
					)]; exists {
						torrents = append(torrents, torrent{
							id:       ts[i].ID,
							name:     ts[i].Name,
							infoHash: InfoHash(ts[i].InfoHash),
						})
						// TODO: make sure download status is updated when
						// fetching metadata
						if dl.Status != DownloadStatusComplete {
							fileStatsRequests <- &torrents[len(torrents)-1]
						}
					}
				}
			}

			return nil
		},
		func(ctx Context) error {
			defer close(updatedTorrents)
			for torrent := range fileStatsRequests {
				if err := ctx.Err(); err != nil {
					return fmt.Errorf(
						"fetching file stats for torrent `%s`: %w",
						torrent.id,
						err,
					)
				}
				files, err := c.Torrents.GetTorrentFileStats(torrent.id)
				if err != nil {
					return fmt.Errorf(
						"fetching file stats for torrent `%s`: %w",
						torrent.id,
						err,
					)
				}
				torrent.files = make([]DownloadFile, len(files))
				for i := range files {
					torrent.files[i] = DownloadFile{
						Path:     files[i].File.Path,
						Progress: int(files[i].BytesCompleted),
						Size:     int(files[i].File.Length),
					}
					torrent.size += uint64(files[i].File.Length)
					torrent.progress += uint64(files[i].BytesCompleted)
				}

				updatedTorrents <- torrent
			}

			return nil
		},
		func(ctx Context) error {
			defer close(updatedDownloads)
			defer close(createTorrentForDownload)
			bestTorrentsByInfoHash := make(map[InfoHash]*torrent)
			for torrent := range updatedTorrents {
				if prev, exists := bestTorrentsByInfoHash[torrent.infoHash]; exists {
					if torrent.progress < prev.progress {
						continue
					}
				}
				bestTorrentsByInfoHash[torrent.infoHash] = torrent
			}

			for i := range downloads {
				torrent, exists := bestTorrentsByInfoHash[downloads[i].InfoHash]
				if !exists {
					createTorrentForDownload <- &downloads[i]
					continue
				}

				downloads[i].Files = torrent.files
				if torrent.progress >= torrent.size {
					downloads[i].Status = DownloadStatusComplete
				}
				updatedDownloads <- &downloads[i]
			}
		},
	)

}

func parallel(ctx Context, tasks ...func(Context) error) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	results := make(chan error)
	for i := range tasks {
		go func() {
			results <- tasks[i](ctx)
		}()
	}

	for range tasks {
		if err := <-results; err != nil {
			return err
		}
	}

	return nil
}
