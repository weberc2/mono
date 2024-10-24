package mm

import (
	"context"
	"errors"
	"fmt"
)

type DownloadStore interface {
	ListDownloads(ctx context.Context) ([]Download, error)
	CreateDownload(ctx context.Context, download *Download) error
	PutDownloads(ctx context.Context, downloads []Download) error
	DeleteDownload(ctx context.Context, infoHash InfoHash) error
}

type DownloadExistsErr struct {
	InfoHash InfoHash `json:"infoHash"`
}

func AsDownloadExistsErr(err error) (e *DownloadExistsErr) {
	errors.As(err, &e)
	return
}

func (err *DownloadExistsErr) Error() string {
	return fmt.Sprintf(
		"download exists for info hash: %s",
		err.InfoHash,
	)
}

type DownloadNotFoundErr struct {
	InfoHash InfoHash `json:"infoHash"`
}

func (err *DownloadNotFoundErr) Error() string {
	return fmt.Sprintf("download not found for info hash: %s", err.InfoHash)
}
