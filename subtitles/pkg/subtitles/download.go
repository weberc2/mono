package subtitles

import (
	"fmt"
	"strings"
	"time"
)

type Download[T Debug] struct {
	ID              DownloadID
	VideoID         T
	Language        string
	OpenSubtitlesID string
	URL             string
	Filepath        string
	Status          DownloadStatus
	Created         time.Time
}

func (d *Download[T]) Debug(sb *strings.Builder) {
	d.VideoID.Debug(sb)
	fmt.Fprintf(
		sb,
		", language=%s, openSubtitlesID=%s, url=%s, filepath=%s, status=%s",
		d.Language,
		d.OpenSubtitlesID,
		d.URL,
		d.Filepath,
		d.Status,
	)
}

type DownloadID string

type DownloadStatus string

const (
	DownloadStatusPending     DownloadStatus = "PENDING"
	DownloadStatusSearching   DownloadStatus = "SEARCHING"
	DownloadStatusFetchingURL DownloadStatus = "FETCHING_URL"
	DownloadStatusDownloading DownloadStatus = "DOWNLOADING"
	DownloadStatusComplete    DownloadStatus = "COMPLETE"
)

type Debug interface {
	Debug(*strings.Builder)
}

func SDebug(d Debug) string {
	var sb strings.Builder
	d.Debug(&sb)
	return sb.String()
}
