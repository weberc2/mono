package main

type Download struct {
	ID             DownloadID     `json:"id,omitempty"`
	Spec           DownloadSpec   `json:"spec"`
	Status         DownloadStatus `json:"status"`
	Torrent        string         `json:"torrent,omitempty"`
	ProcessedFiles []string       `json:"processedFiles,omitempty"`
	Error          string         `json:"error,omitempty"`
}

type DownloadID string

type DownloadStatus string

const (
	DownloadStatusPending          DownloadStatus = "PENDING"
	DownloadStatusFetchingMetadata DownloadStatus = "FETCHING_METADATA"
	DownloadStatusVerifying        DownloadStatus = "VERIFYING"
	DownloadStatusDownloading      DownloadStatus = "DOWNLOADING"
	DownloadStatusPostProcessing   DownloadStatus = "POST_PROCESSING"
	DownloadStatusSuccess          DownloadStatus = "SUCCESS"
	DownloadStatusError            DownloadStatus = "ERROR"
)

type DownloadSpec struct {
	// Source describes how to acquire the content.
	Source Source `json:"source"`

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
	// Path is the path to the file.
	Path string `json:"path"`

	// Kind is the kind of the file (e.g., video or subtitle)
	Kind FileKind `json:"kind"`

	// Type is the type of the media.
	Type MediaType `json:"type"`

	// Title is the title of the film or show.
	Title string `json:"title"`

	// Year is the year the film or show was released.
	Year string `json:"year"`

	// Season is the season of the show. Not applicable for films.
	Season string `json:"season,omitempty"`

	// Episode is the episode of the show. Not applicable for films.
	Episode string `json:"episode,omitempty"`

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
