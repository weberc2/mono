package main

import "github.com/anacrolix/torrent/metainfo"

type Download struct {
	ID          DownloadID        `json:"id"`
	Labels      []string          `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	Spec        DownloadSpec      `json:"spec"`
	Status      DownloadStatus    `json:"status"`
	Error       string            `json:"error,omitempty"`
	Files       []FileDownload    `json:"files,omitempty"`
}

type FileDownload struct {
	Path            string             `json:"path"`
	Status          FileDownloadStatus `json:"status"`
	ProgressPercent uint               `json:"progressPercent"`
}

type FileDownloadStatus string

const (
	FileDownloadStatusPending  FileDownloadStatus = "PENDING"
	FileDownloadStatusProgress FileDownloadStatus = "PROGRESS"
)

type DownloadID string

type DownloadStatus string

const (
	DownloadStatusPending          DownloadStatus = "PENDING"
	DownloadStatusFetchingMetadata DownloadStatus = "FETCHING_METADATA"
	DownloadStatusDownloading      DownloadStatus = "DOWNLOADING"
	DownloadStatusSuccess          DownloadStatus = "SUCCESS"
	DownloadStatusFailure          DownloadStatus = "FAILURE"
)

type DownloadSpec struct {
	// Type is the type of the download.
	Type DownloadType `json:"type"`

	// MetaInfo is the structured representation of a torrent file.
	MetaInfo metainfo.MetaInfo `json:"metaInfo,omitempty"`

	// Torrent is the content of a torrent file.
	Torrent string `json:"torrent,omitempty"`

	// URL is the URL for a `MAGNET` or `HTTP` source.
	URL URL `json:"url,omitempty"`
}

type DownloadType string

const (
	DownloadTypeMetaInfo DownloadType = "METAINFO"
	DownloadTypeTorrent  DownloadType = "TORRENT"
	DownloadTypeMagnet   DownloadType = "MAGNET"
	DownloadTypeHTTP     DownloadType = "HTTP"
)
