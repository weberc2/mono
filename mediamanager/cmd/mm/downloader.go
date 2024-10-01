package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cenkalti/rain/rainrpc"
)

type Downloader struct {
	Client   *rainrpc.Client
	Trackers []string
	cache    cache
}

var _ DownloadStore = (*Downloader)(nil)

func NewDownloader(
	client *rainrpc.Client,
	cacheTTL time.Duration,
	trackers ...string,
) (downloader Downloader) {
	downloader.Client = client
	downloader.Trackers = trackers
	downloader.cache = cache{
		entries: make(map[string]*cacheEntry),
		ttl:     cacheTTL,
	}
	return
}

func (d *Downloader) CreateDownload(
	ctx context.Context,
	dl *Download,
) (download Download, err error) {
	if err = ctx.Err(); err != nil {
		err = fmt.Errorf("creating download: %w", err)
		return
	}

	if _, err = d.Client.AddURI(
		MagnetLink(dl.Name, dl.InfoHash, d.Trackers...),
		&rainrpc.AddTorrentOptions{ID: string(dl.InfoHash)},
	); err != nil {
		err = fmt.Errorf("creating download: %w", err)
		return
	}

	download = *dl
	download.Status = DownloadStatusPending
	return
}

func (d *Downloader) FetchDownload(
	ctx context.Context,
	infoHash InfoHash,
) (download Download, err error) {
	if err = ctx.Err(); err != nil {
		err = fmt.Errorf("fetching download `%s`: %w", infoHash, err)
		return
	}

	stats, e := d.Client.GetTorrentStats(string(infoHash))
	if e != nil {
		err = fmt.Errorf("fetching download `%s`: %w", infoHash, e)
		return
	}

	download.InfoHash = InfoHash(stats.InfoHash)
	if stats.Status == "Downloading Metadata" {
		download.Status = DownloadStatusFetchingMetadata
		return
	} else if stats.Bytes.Downloaded >= stats.Bytes.Total {
		download.Status = DownloadStatusComplete
	} else {
		download.Status = DownloadStatusProgress
	}

	if download.Files, err = d.cache.fetch(
		ctx,
		d.Client,
		string(infoHash),
	); err != nil {
		err = fmt.Errorf(
			"fetching download `%s`: fetching file stats: %w",
			infoHash,
			err,
		)
		return
	}
	return
}

func (d *Downloader) ListDownloads(
	ctx context.Context,
) ([]Download, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("listing downloads: %w", err)
	}

	torrents, err := d.Client.ListTorrents()
	if err != nil {
		return nil, fmt.Errorf("listing downloads: %w", err)
	}

	// TODO (perf): instead of a cache on the file-stats endpoint, we could just
	// run goroutines that poll constantly and keep the file-stats cache hot all
	// the time so we always have instantaneous file-stats (for that matter, we
	// could just keep all of the download information synced all the time).
	downloads := make([]Download, len(torrents))
	for i := range torrents {
		downloads[i], err = d.FetchDownload(ctx, InfoHash(torrents[i].ID))
		if err != nil {
			return nil, fmt.Errorf("listing downloads: %w", err)
		}
	}

	return downloads, nil
}

type cache struct {
	entries map[string]*cacheEntry
	lock    sync.Mutex
	ttl     time.Duration
}

func (c *cache) fetch(
	ctx context.Context,
	client *rainrpc.Client,
	id string,
) ([]DownloadFile, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	entry, exists := c.entries[id]
	if exists && time.Now().Before(entry.ttl) {
		return entry.files, nil
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf(
			"fetching file stats for torrent `%s`: %w",
			id,
			err,
		)
	}

	stats, err := client.GetTorrentFileStats(id)
	if err != nil {
		return nil, err
	}

	files := make([]DownloadFile, len(stats))
	for i := range stats {
		files[i] = DownloadFile{
			Path:     stats[i].File.Path,
			Size:     int(stats[i].File.Length),
			Progress: int(stats[i].BytesCompleted),
		}
	}

	c.entries[id] = &cacheEntry{ttl: time.Now().Add(c.ttl), files: files}
	return files, nil
}

type cacheEntry struct {
	ttl   time.Time
	files []DownloadFile
}
