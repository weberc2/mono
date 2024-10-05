package main

// FilmTransform defines a transform for a film download.
type FilmTransform struct {
	// Title is the title of the film.
	Title string `json:"title"`

	// Year is the year the film was released.
	Year string `json:"year"`

	// Files are the files to import into the media library.
	Files []MediaFile `json:"files"`
}

// MediaFile defines a file inside of a media download.
type MediaFile struct {
	// Path is the path to the file in the download.
	Path string `json:"path"`

	// Type is the type of the media file (`VIDEO` or `SUBTITLE`).
	Type MediaFileType `json:"type"`

	// Language is the language for subtitle files.
	Language string `json:"language,omitempty"`
}

// MediaFileType is an enumeration of types of media files.
type MediaFileType string

const (
	MediaFileTypeVideo    MediaFileType = "VIDEO"
	MediaFileTypeSubtitle MediaFileType = "SUBTITLE"
)
