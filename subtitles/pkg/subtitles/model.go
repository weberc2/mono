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
	if mf.Kind == MediaFileKindVideo {

	}
	slog.Info(
		"inserting media file",
		"title", mf.Title,
		"season", mf.Season,
		"episode", mf.Episode,
		"kind", mf.Kind,
	)
	return nil
}

func (m Model) UpsertShowVideoFile(
	filepath string,
)

const upsertShowVideoFilesQuery = `INSERT INTO showvideofiles (
	filepath,
	title,
	season,
	episode,
	mediahash
) VALUES ($1, $2, $3, $4, $5)
ON CONFLICT DO UPDATE
SET title=$2, season=$3, episode=$4, mediahash=$5;`
