package subtitles

type ShowVideoFile struct {
	ShowMediaFile
	Mediahash string `json:"mediahash"`
}

type ShowSubtitleFile struct {
	ShowMediaFile
	Language string `json:"language"`
}

type ShowMediaFile struct {
	Filepath string `json:"filepath"`
	Title string `json:"title"`
	Season string `json:"season"`
	Episode string `json:"episode"`
}

func (file *ShowMediaFile) Subtitle() (sub ShowSubtitleFile) {
	sub = ShowSubtitleFile{
		Filepath: file.Filepath,
		Title: file.Title,
		Season: file.Season,
		Episode: file.Episode,
		Language:
	}
}

type MediaFile struct {
	// Filepath is the path to the media file.
	Filepath string `json:"filepath"`

	// Type is the type of the media file. This determines what fields are set.
	Type MediaType `json:"type"`

	// Kind is the kind of the media file--whether it's a video or a subtitle
	// file.
	Kind MediaFileKind `json:"kind"`

	// Title is the title for a film or show. This is required for all media
	// files.
	Title string `json:"title"`

	// Season is the season for a show. This is required whether the media file
	// is a show subtitle or a show video file, but must not be set for film media.
	Season string `json:"season,omitempty"`

	// Episode is the episode number for a show. This is required for show
	// subtitle and show video files, but must not be set for film media.
	Episode string `json:"episode,omitempty"`

	// Mediahash is the mediahash for video media. Must not be set for subtitle
	// media.
	Mediahash string `json:"mediahash,omitempty"`

	// SubtitleLanguage is the language for a subtitle media type. Should be
	// empty for video media files.
	SubtitleLanguage string `json:"language,omitempty"`

	// Framerate is the framerate for a subtitle or video media file.
	Framerate uint8 `json:"framerate"`
}

type MediaType string

const (
	MediaTypeFilm          MediaType = "FILM"
	MediaTypeShow          MediaType = "SHOW"
	MediaTypeExtra         MediaType = "EXTRA"
	MediaTypeDeletedScenes MediaType = "DELETEDSCENES"
)

type MediaFileKind string

const (
	MediaFileKindSubtitle MediaFileKind = "SUBTITLE"
	MediaFileKindVideo    MediaFileKind = "VIDEO"
)
