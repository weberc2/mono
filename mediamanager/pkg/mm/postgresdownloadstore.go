package mm

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresDownloadStore struct {
	DB *pgxpool.Pool
}

var _ DownloadStore = PostgresDownloadStore{}

func (store PostgresDownloadStore) ListDownloads(
	ctx context.Context,
) (downloads []Download, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("listing downloads: %w", err)
		}
	}()

	var rows pgx.Rows
	if rows, err = store.DB.Query(ctx, listDownloadsQuery); err != nil {
		return
	}
	defer rows.Close()

	var last Download
	for rows.Next() {
		var row Download
		var path sql.NullString
		var size sql.NullInt64
		var progress sql.NullInt64
		if err = rows.Scan(
			&row.ID,
			&row.Status,
			&row.Size,
			&row.Progress,
			&path,
			&size,
			&progress,
		); err != nil {
			err = fmt.Errorf("scanning into downloads: %w", err)
			return
		}

		// if the download identifier has changed, it means we're no longer
		// scanning in a new file for an existing download, but rather that
		// we're on an altogether new download. append it to the output slice.
		if row.ID != last.ID {
			downloads = append(downloads, row)
			last = row
		}

		// in theory, a download could have no files, in which case the `LEFT
		// JOIN` would return a row with `NULL` for the file fields. Because
		// these fields are all `NOT NULL` in the database, these should only
		// ever be `NULL` as a result of this no-files-for-a-given-download
		// `LEFT JOIN` behavior. they should also only ever be all null or all
		// valid, but we'll check all of them here anyway.
		if path.Valid && size.Valid && progress.Valid {
			d := &downloads[len(downloads)-1]
			d.Files = append(
				d.Files,
				DownloadFile{
					Path:     path.String,
					Size:     uint64(size.Int64),
					Progress: uint64(progress.Int64),
				},
			)
		}
	}

	return
}

const listDownloadsQuery = `
SELECT
	d.id, d.status, d.size, d.progress, f.path, f.size, f.progress
FROM downloads d
LEFT JOIN downloadfiles f ON d.id = f.download;`

func (store PostgresDownloadStore) FetchDownload(
	ctx context.Context,
	infoHash InfoHash,
) (download Download, err error) {
	var filesJSON []byte
	if err = store.DB.QueryRow(
		ctx,
		fetchDownloadQuery,
		infoHash,
	).Scan(
		&download.ID,
		&download.Status,
		&download.Size,
		&download.Progress,
		&filesJSON,
	); err != nil {
		err = fmt.Errorf(
			"fetching download: scanning row: %w",
			err,
		)
		return
	}

	if err = json.Unmarshal(filesJSON, &download.Files); err != nil {
		err = fmt.Errorf(
			"fetching download: deserializing download files: %w",
			err,
		)
	}
	return
}

const fetchDownloadQuery = `
SELECT
	d.id,
	d.status,
	d.size,
	d.progress,
	JSONB_AGG(JSONB_BUILD_OBJECT(
		'path', f.path,
		'size', f.size,
		'progress', f.progress)) AS files
FROM downloads d
LEFT JOIN downloadfiles f ON d.id = f.download
WHERE d.id = $1
GROUP BY d.id;`

func (store PostgresDownloadStore) CreateDownload(
	ctx context.Context,
	download *Download,
) (err error) {
	defer func() {
		if err != nil {
			if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
				err = &DownloadExistsErr{InfoHash: download.ID}
			}
			err = fmt.Errorf("creating download `%s`: %w", download.ID, err)
		}
	}()

	var tx pgx.Tx
	if tx, err = store.DB.Begin(ctx); err != nil {
		err = fmt.Errorf("beginning transaction: %w", err)
		return
	}

	var success bool
	defer func() {
		if success {
			if e := tx.Commit(ctx); e != nil {
				err = errors.Join(
					err,
					fmt.Errorf("committing transaction: %w", err),
				)
			}
		} else {
			if e := tx.Rollback(ctx); e != nil {
				err = errors.Join(
					err,
					fmt.Errorf("rolling back transaction: %w", err),
				)
			}
		}
	}()

	var batch pgx.Batch
	batch.Queue(
		createDownloadQuery0,
		download.ID,
		download.Status,
		download.Size,
		download.Progress,
	)

	for i := range download.Files {
		batch.Queue(
			createDownloadQuery1,
			download.ID,
			download.Files[i].Path,
			download.Files[i].Size,
			download.Files[i].Progress,
		)
	}

	results := tx.SendBatch(ctx, &batch)
	defer func() {
		if e := results.Close(); e != nil {
			err = errors.Join(err, fmt.Errorf("closing batch results: %w", e))
		}
	}()

	for range batch.QueuedQueries {
		if _, err = results.Exec(); err != nil {
			return
		}
	}

	success = true // commit tx rather than rolling it back
	return
}

const createDownloadQuery0 = `
INSERT INTO downloads (id, status, size, progress) VALUES($1, $2, $3, $4);`

const createDownloadQuery1 = `
INSERT INTO downloadfiles (download, path, size, progress)
VALUES($1, $2, $3, $4);`

func (store PostgresDownloadStore) PutDownloads(
	ctx context.Context,
	downloads []Download,
) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("putting downloads: %w", err)
		}
	}()

	var tx pgx.Tx
	if tx, err = store.DB.Begin(ctx); err != nil {
		err = fmt.Errorf("beginning transaction: %w", err)
		return
	}

	var success bool
	defer func() {
		if success {
			if e := tx.Commit(ctx); e != nil {
				err = errors.Join(
					err,
					fmt.Errorf("committing transaction: %w", err),
				)
			}
		} else {
			if e := tx.Rollback(ctx); e != nil {
				err = errors.Join(
					err,
					fmt.Errorf("rolling back transaction: %w", err),
				)
			}
		}
	}()

	var batch pgx.Batch
	for i := range downloads {
		// upsert the download
		batch.Queue(
			putDownloadsQuery0,
			downloads[i].ID,
			downloads[i].Status,
			downloads[i].Size,
			downloads[i].Progress,
		)

		// delete any existing files associated with the download
		batch.Queue(putDownloadsQuery1, downloads[i].ID)

		// replace them with any new files associated with the updated download.
		for j := range downloads[i].Files {
			batch.Queue(
				putDownloadsQuery2,
				downloads[i].ID,
				downloads[i].Files[j].Path,
				downloads[i].Files[j].Size,
				downloads[i].Files[j].Progress,
			)
		}
	}

	results := tx.SendBatch(ctx, &batch)
	defer func() {
		if e := results.Close(); e != nil {
			err = errors.Join(err, fmt.Errorf("closing batch results: %w", err))
		}
	}()

	for range batch.QueuedQueries {
		if _, err = results.Exec(); err != nil {
			return
		}
	}

	success = true // commit tx rather than rolling it back
	return
}

const putDownloadsQuery0 = `
INSERT INTO downloads (id, status, size, progress) VALUES($1, $2, $3, $4)
ON CONFLICT (id) DO UPDATE SET status=$2, size=$3, progress=$4;`

const putDownloadsQuery1 = `
DELETE FROM downloadfiles WHERE download=$1;`

const putDownloadsQuery2 = `
INSERT INTO downloadfiles (download, path, size, progress)
VALUES($1, $2, $3, $4)`

func (store PostgresDownloadStore) DeleteDownload(
	ctx context.Context,
	infoHash InfoHash,
) (err error) {
	var sentinel int
	if err = store.DB.QueryRow(
		ctx,
		deleteDownloadQuery,
		infoHash,
	).Scan(&sentinel); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = fmt.Errorf(
				"deleting download: %w",
				&DownloadNotFoundErr{InfoHash: infoHash},
			)
		} else {
			err = fmt.Errorf("deleting download `%s`: %w", infoHash, err)
		}
	}
	return
}

const deleteDownloadQuery = `DELETE FROM downloads WHERE id=$1 RETURNING 1;`
