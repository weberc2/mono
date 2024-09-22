package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
)

func main() {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "./config.json"
	}
	if err := RunFromPath(context.Background(), path); err != nil {
		log.Fatal(err)
	}
}

type downloadStore struct {
	lock      sync.RWMutex
	downloads map[DownloadID]*Download
}

func newDownloadStore() *downloadStore {
	return &downloadStore{downloads: make(map[DownloadID]*Download)}
}

var _ DownloadStore = (*downloadStore)(nil)

func (store *downloadStore) CreateDownload(
	ctx Context,
	d *Download,
) (download Download, err error) {
	store.lock.Lock()
	defer store.lock.Unlock()
	if _, exists := store.downloads[d.ID]; exists {
		err = fmt.Errorf(
			"creating media download `%s`: %w",
			d.ID,
			DownloadExistsErr{d.ID},
		)
		return
	}
	download = *d
	download.Status = DownloadStatusPending
	store.downloads[d.ID] = &download
	return
}

func (store *downloadStore) ListDownloads(
	ctx Context,
) (downloads []Download, err error) {
	store.lock.RLock()
	defer store.lock.RUnlock()
	for _, download := range store.downloads {
		downloads = append(downloads, *download)
	}
	return
}

func (store *downloadStore) UpdateDownloadStatus(
	ctx Context,
	id DownloadID,
	status DownloadStatus,
	torrent string,
	processedFiles []string,
	errorMessage string,
) (download Download, err error) {
	md, exists := store.downloads[id]
	if !exists {
		err = fmt.Errorf(
			"updating status for media download `%s`: not found",
			id,
		)
		return
	}
	md.Status = status
	md.Torrent = torrent
	md.ProcessedFiles = processedFiles
	md.Error = errorMessage
	download = *md
	return
}
