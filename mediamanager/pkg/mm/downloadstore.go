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
}

type DownloadExistsErr struct {
	InfoHash InfoHash
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
