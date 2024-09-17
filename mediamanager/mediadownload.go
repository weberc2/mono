package main

type MediaDownload struct {
	ID             MediaDownloadID     `json:"id,omitempty"`
	Spec           MediaDownloadSpec   `json:"spec"`
	Status         MediaDownloadStatus `json:"status"`
	Download       DownloadID          `json:"download,omitempty"`
	Transformation TransformationID    `json:"transformation,omitempty"`
	Error          string              `json:"error,omitempty"`
}

type MediaDownloadID string

type MediaDownloadStatus string

const (
	MediaDownloadStatusPending          MediaDownloadStatus = "PENDING"
	MediaDownloadStatusFetchingMetadata MediaDownloadStatus = "FETCHING_METADATA"
	MediaDownloadStatusValidating       MediaDownloadStatus = "VALIDATING"
	MediaDownloadStatusDownloading      MediaDownloadStatus = "DOWNLOADING"
	MediaDownloadStatusDownloadComplete MediaDownloadStatus = "DOWNLOAD_COMPLETE"
	MediaDownloadStatusProcessing       MediaDownloadStatus = "PROCESSING"
	MediaDownloadStatusSuccess          MediaDownloadStatus = "SUCCESS"
	MediaDownloadStatusFailure          MediaDownloadStatus = "FAILURE"
)

type MediaDownloadSpec struct {
	// Download describes how to acquire the content.
	Download DownloadSpec `json:"download"`

	// Files contains the information about the files in the download that the
	// application should handle.
	Files MediaFiles `json:"files"`
}

type MediaFiles struct {
	Type MediaFilesType `json:"type"`
	List []MediaFile    `json:"list,omitempty"`
}

type MediaFilesType string

const (
	MediaFilesTypeList  MediaFilesType = "LIST"
	MediaFilesTypeInfer MediaFilesType = "INFER"
)

type MediaFile struct {
	// Kind is the kind of the file (e.g., video or subtitle)
	Kind FileKind `json:"kind"`

	// Type is the type of the media.
	Type MediaType `json:"type"`

	// Path is the path to the file.
	Path string `json:"path"`

	// Lang is the language for subtitle files.
	Lang string `json:"lang,omitempty"`
}

type MediaType string

const (
	MediaTypeFilm    MediaType = "FILM"
	MediaTypeEpisode MediaType = "EPISODE"
)

type FileKind string

const (
	FileKindVideo    FileKind = "VIDEO"
	FileKindSubtitle FileKind = "SUBTITLE"
)
