package subtitles

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/lib/pq"
)

type Model struct {
	DB *sql.DB
}

func (m Model) TryReserveDownloads(
	ctx context.Context,
	count int,
) (token Reservation, downloads []Download[Episode], err error) {
	slog.Debug(
		"model: trying to reserve downloads",
		"count", count,
	)

	// this query reserves all downloads which have not been completed AND which
	// are not already reserved. a download is considered "reserved" if fewer
	// than $TTL seconds have transpired since the download's `lastreserved`
	// time. the mechanism by which this query reserves the download is by
	// setting its `lastreserved` field to the current timestamp.
	const query = `UPDATE showdownloads SET lastreserved=NOW()
		WHERE title, year, season, episode, language IN (
			SELECT FROM showdownloads
			WHERE lastreserved + 10s < NOW()
			AND status != 'COMPLETED'
			LIMIT $1
		)
	RETURNING
		title,
		year,
		season,
		episode,
		language,
		opensubtitlesid,
		url,
		filepath,
		status,
		created;`

	var rows *sql.Rows
	if rows, err = m.DB.QueryContext(
		ctx,
		query,
		count,
	); err != nil {
		err = fmt.Errorf("reserving downloads: %w", err)
		return
	}

	var t time.Time
	downloads = make([]Download[Episode], 0, count)
	for rows.Next() {
		downloads = append(downloads, Download[Episode]{})
		d := &downloads[len(downloads)-1]
		if err = rows.Scan(
			&d.ID,
			&d.VideoID.Title,
			&d.VideoID.Year,
			&d.VideoID.Season,
			&d.VideoID.Episode,
			&d.Language,
			&d.OpenSubtitlesID,
			&d.URL,
			&d.Filepath,
			&d.Status,
			&d.Created,
			&t,
		); err != nil {
			err = fmt.Errorf("reserving downloads: scanning row: %w", err)
			return
		}
	}
	token = Reservation(t)
	return
}

func (m Model) SetOpenSubtitlesID(
	ctx context.Context,
	reservation Reservation,
	episode *Episode,
	language string,
	openSubtitlesID string,
) (download Download[Episode], err error) {
	slog.Debug(
		"model: updating download with OpenSubtitles ID",
		"title", episode.Title,
		"year", episode.Year,
		"season", episode.Season,
		"episode", episode.Episode,
		"language", language,
		"openSubtitlesID", openSubtitlesID,
	)

	if err = reservation.Valid(); err != nil {
		err = fmt.Errorf(
			"setting opensubtitlesid for `%s (%s): %w",
			SDebug(episode),
			language,
			err,
		)
		return
	}

	const query = `UPDATE showdownloads SET opensubtitlesid=$6`
}

func (m Model) InsertPendingEpisodeDownload(
	ctx context.Context,
	episode *Episode,
	language string,
) (download Download[Episode], err error) {
	slog.Debug(
		"model: inserting pending episode subtitle download",
		"title", episode.Title,
		"year", episode.Year,
		"season", episode.Season,
		"episode", episode.Episode,
		"language", language,
	)

	const query = `INSERT INTO showdownloads (
		title,           -- 1
		year,            -- 2
		season,          -- 3
		episode,         -- 4
		language,        -- 5
		opensubtitlesid, -- 6
		url,             -- 7
		filepath,        -- 8
		status           -- 9
		created          -- 10
		lastreserved     -- 11
	) VALUES ($1, $2, $3, $4, $5, '', '', '', 'PENDING', NOW(), NULL)
	RETURNING
		title,
		year,
		season,
		episode,
		language,
		opensubtitlesid,
		url,
		filepath,
		status,
		created;`

	if err = m.DB.QueryRowContext(
		ctx,
		query,
		episode.Title,
		episode.Year,
		episode.Season,
		episode.Episode,
		language,
	).Scan(
		&download.VideoID.Title,
		&download.VideoID.Year,
		&download.VideoID.Season,
		&download.VideoID.Episode,
		&download.Language,
		&download.OpenSubtitlesID,
		&download.URL,
		&download.Filepath,
		&download.Status,
		&download.Created,
	); err != nil {
		// TODO: handle exists
		err = fmt.Errorf(
			"inserting download `%s (%s)`: %w",
			SDebug(episode),
			language,
			err,
		)
	}
	return
}

func (m Model) InsertEpisodeVideoFile(
	ctx context.Context,
	f *VideoFile[Episode],
) error {
	slog.Debug(
		"inserting episode video file",
		"title", f.ID.Title,
		"year", f.ID.Year,
		"season", f.ID.Season,
		"episode", f.ID.Episode,
		"filepath", f.Filepath,
	)

	const query = `INSERT INTO showvideofiles (
		filepath, title, year, season, episode
	) VALUES($1, $2, $3, $4, $5)`

	if _, err := m.DB.ExecContext(
		ctx,
		query,
		f.Filepath,
		f.ID.Title,
		f.ID.Year,
		f.ID.Season,
		f.ID.Episode,
	); err != nil {
		return fmt.Errorf(
			"inserting video file `%s` for show `%s (%s) Season %s Episode "+
				"%s: %w",
			f.Filepath,
			f.ID.Title,
			f.ID.Year,
			f.ID.Season,
			f.ID.Episode,
			err,
		)
	}
	return nil
}

func (m Model) InsertEpisodeSubtitleFile(
	ctx context.Context,
	f *SubtitleFile[Episode],
) error {
	slog.Info(
		"inserting show subtitle file",
		"title", f.ID.Title,
		"year", f.ID.Year,
		"season", f.ID.Season,
		"episode", f.ID.Episode,
		"language", f.Language,
		"filepath", f.Filepath,
	)
	const query = `INSERT INTO showsubtitlefiles (
		filepath, title, year, season, episode, language
	) VALUES($1, $2, $3, $4, $5, $6)`

	if _, err := m.DB.ExecContext(
		ctx,
		query,
		f.Filepath,
		f.ID.Title,
		f.ID.Year,
		f.ID.Season,
		f.ID.Episode,
		f.Language,
	); err != nil {
		return fmt.Errorf(
			"inserting subtitle file `%s` for episode `%s (%s) Season %s "+
				"Episode %s (%s): %w",
			f.Filepath,
			f.ID.Title,
			f.ID.Year,
			f.ID.Season,
			f.ID.Episode,
			f.Language,
			err,
		)
	}
	return nil
}

type Reservation time.Time

func (r Reservation) Valid() error {
	if time.Now().Before(time.Time(r)) {
		return nil
	}
	return ReservationExpiredErr(r)
}

type ReservationExpiredErr Reservation

func (err ReservationExpiredErr) Error() string {
	return fmt.Sprintf("reservation expired at `%s`", time.Time(err))
}

type DownloadExistsErr[T Debug] struct {
	VideoID  T
	Language string
}

func (err *DownloadExistsErr[T]) Error() string {
	return fmt.Sprintf(
		"download exists: `%s (%s)`",
		SDebug(err.VideoID),
		err.Language,
	)
}
