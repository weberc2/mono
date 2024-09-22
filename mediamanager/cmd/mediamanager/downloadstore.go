package main

import (
	"fmt"
	"net/http"
)

type DownloadStore interface {
	ListDownloads(ctx Context) (downloads []Download, err error)
	CreateDownload(ctx Context, d *Download) (download Download, err error)
	UpdateDownloadStatus(
		ctx Context,
		id DownloadID,
		status DownloadStatus,
		torrent string,
		processedFiles []string,
		errorMessage string,
	) (download Download, err error)
}

type DownloadExistsErr struct {
	Download DownloadID `json:"download"`
}

func (err *DownloadExistsErr) Error() string {
	return fmt.Sprintf("download exists: %s", err.Download)
}

func (err *DownloadExistsErr) Details() HTTPErrorDetails {
	return HTTPErrorDetails{
		Status:  http.StatusConflict,
		Message: "Download Exists",
		Details: err,
	}
}
