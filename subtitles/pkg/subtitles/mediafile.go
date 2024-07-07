package subtitles

type VideoFile[T any] struct {
	Filepath string
	ID       T
}

type SubtitleFile[T any] struct {
	VideoFile[T]
	Language string
}

type MediaFile[T any] struct {
	SubtitleFile[T]
	IsSubtitle bool
}
