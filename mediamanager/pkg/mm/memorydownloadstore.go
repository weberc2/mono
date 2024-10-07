package mm

import (
	"context"
	"fmt"
	"sync"
)

type MemoryDownloadStore struct {
	lock      sync.RWMutex
	downloads []Download
}

var _ DownloadStore = (*MemoryDownloadStore)(nil)

func (store *MemoryDownloadStore) ListDownloads(
	ctx context.Context,
) (downloads []Download, err error) {
	store.lock.RLock()
	defer store.lock.RUnlock()

	downloads = make([]Download, len(store.downloads))
	copy(downloads, store.downloads)

	for i := range downloads {
		copy(downloads[i].Files, store.downloads[i].Files)
	}

	return
}

func (store *MemoryDownloadStore) CreateDownload(
	ctx context.Context,
	download *Download,
) error {
	store.lock.Lock()
	defer store.lock.Unlock()

	for i := range store.downloads {
		if store.downloads[i].ID == download.ID {
			return fmt.Errorf(
				"creating download: %w",
				&DownloadExistsErr{InfoHash: download.ID},
			)
		}
	}

	store.downloads = append(store.downloads, *download)
	return nil
}

func (store *MemoryDownloadStore) PutDownload(
	ctx context.Context,
	download *Download,
) error {
	store.lock.Lock()
	defer store.lock.Unlock()

	for i := range store.downloads {
		if store.downloads[i].ID == download.ID {
			store.downloads[i] = *download
			return nil
		}
	}
	store.downloads = append(store.downloads, *download)
	return nil
}

func (store *MemoryDownloadStore) PutDownloads(
	ctx context.Context,
	downloads []Download,
) error {
	for i := range downloads {
		// cannot error
		_ = store.PutDownload(ctx, &downloads[i])
	}
	return nil
}
