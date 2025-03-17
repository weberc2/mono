package subtitles

import (
	"fmt"
	"strings"
	"time"
)

type Download[T Debug] struct {
	ID              DownloadID[T]
	OpenSubtitlesID string
	URL             string
	Filepath        string
	Status          DownloadStatus
	Created         time.Time
}

func (d *Download[T]) Debug(sb *strings.Builder) {
	d.ID.Debug(sb)
	fmt.Fprintf(
		sb,
		", openSubtitlesID=%s, url=%s, filepath=%s, status=%s",
		d.OpenSubtitlesID,
		d.URL,
		d.Filepath,
		d.Status,
	)
}

type DownloadID[T Debug] struct {
	Video    T
	Language string
}

func (d DownloadID[T]) Debug(sb *strings.Builder) {
	d.Video.Debug(sb)
	fmt.Fprintf(sb, ", language=%s", d.Language)
}

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
