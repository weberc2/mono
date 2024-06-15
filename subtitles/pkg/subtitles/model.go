package subtitles

import (
	"database/sql"
	"log/slog"
)

type Model struct {
	DB *sql.DB
}

func (m Model) ListMediaFilesByType(mediaType MediaType) ([]MediaFile, error) {
	panic("Model.ListMediaFiles() not implemented")
}

func (m Model) InsertMediaFile(mf *MediaFile) error {
	slog.Info(
		"inserting media file",
		"title", mf.Title,
		"season", mf.Season,
		"episode", mf.Episode,
		"kind", mf.Kind,
	)
	return nil
}
