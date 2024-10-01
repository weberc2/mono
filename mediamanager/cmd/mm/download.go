package main

type Download struct {
	Name     string         `json:"name"`
	InfoHash InfoHash       `json:"infoHash"`
	Status   DownloadStatus `json:"status"`
	Files    []DownloadFile `json:"files"`
}

type DownloadFile struct {
	// Path is the path to the file.
	Path string `json:"path"`

	// Size is the size of the file in bytes.
	Size int `json:"size"`

	// Progress is the number of bytes downloaded.
	Progress int `json:"progress"`
}

type DownloadStatus string

const (
	DownloadStatusPending          DownloadStatus = "PENDING"
	DownloadStatusFetchingMetadata DownloadStatus = "FETCHING-METADATA"
	DownloadStatusProgress         DownloadStatus = "PROGRESS"
	DownloadStatusComplete         DownloadStatus = "COMPLETE"
)

type InfoHash string
