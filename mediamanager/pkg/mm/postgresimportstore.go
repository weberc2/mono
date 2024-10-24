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

type PostgresImportStore struct {
	DB *pgxpool.Pool
}

var _ ImportStore = PostgresImportStore{}

func (store PostgresImportStore) ListImports(
	ctx context.Context,
) (imports []Import, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("listing imports: %w", err)
		}
	}()

	var rows pgx.Rows
	if rows, err = store.DB.Query(ctx, listImportsQuery); err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		imports = append(imports, Import{})
		row := &imports[len(imports)-1]

		var spec, status json.RawMessage
		if err = rows.Scan(
			&row.ID,
			&spec,
			&status,
		); err != nil {
			return
		}

		if err = json.Unmarshal(spec, row); err != nil {
			err = fmt.Errorf("unmarshaling spec: %w", err)
			return
		}

		if err = json.Unmarshal(status, row); err != nil {
			err = fmt.Errorf("unmarshaling status: %w", err)
			return
		}
	}

	return
}

const listImportsQuery = `SELECT id, spec, status FROM imports;`

func (store PostgresImportStore) CreateImport(
	ctx context.Context,
	imp *Import,
) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("creating import: %w", err)
		}
	}()

	var spec, status []byte

	if spec, err = json.Marshal(struct {
		InfoHash InfoHash `json:"infoHash"`
		Film     *Film    `json:"film,omitempty"`
	}{
		InfoHash: imp.InfoHash,
		Film:     imp.Film,
	}); err != nil {
		err = fmt.Errorf("marshaling spec: %w", err)
		return
	}

	if status, err = json.Marshal(struct {
		Status ImportStatus `json:"status"`
		Files  ImportFiles  `json:"files"`
	}{
		Status: imp.Status,
		Files:  imp.Files,
	}); err != nil {
		err = fmt.Errorf("marshaling status: %w", err)
		return
	}

	if _, err = store.DB.Exec(
		ctx,
		createImportQuery,
		imp.ID,
		spec,
		status,
	); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			err = &ImportExistsErr{Import: imp.ID}
		}
	}
	return
}

const createImportQuery = `
INSERT INTO imports (id, spec, status) VALUES($1, $2, $3);`

func (store PostgresImportStore) DeleteImport(
	ctx context.Context,
	id ImportID,
) error {
	var sentinel int
	if err := store.DB.QueryRow(
		ctx,
		deleteImportQuery,
		id,
	).Scan(&sentinel); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf(
				"deleting import: %w",
				&ImportNotFoundErr{Import: id},
			)
		}
		return fmt.Errorf("deleting import `%s`: %w", id, err)
	}
	return nil
}

const deleteImportQuery = "DELETE FROM imports WHERE id=$1 RETURNING 1;"

func (store PostgresImportStore) UpdateImport(
	ctx context.Context,
	imp *Import,
) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("updating import: %w", err)
		}
	}()

	var spec, status []byte

	if spec, err = json.Marshal(struct {
		InfoHash InfoHash `json:"infoHash"`
		Film     *Film    `json:"film,omitempty"`
	}{
		InfoHash: imp.InfoHash,
		Film:     imp.Film,
	}); err != nil {
		err = fmt.Errorf("marshaling spec: %w", err)
		return
	}

	if status, err = json.Marshal(struct {
		Status ImportStatus `json:"status"`
		Files  ImportFiles  `json:"files"`
	}{
		Status: imp.Status,
		Files:  imp.Files,
	}); err != nil {
		err = fmt.Errorf("marshaling status: %w", err)
		return
	}

	var sentinel int
	if err = store.DB.QueryRow(
		ctx,
		updateImportQuery,
		imp.ID,
		spec,
		status,
	).Scan(&sentinel); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = &ImportNotFoundErr{Import: imp.ID}
		}
	}
	return
}

const updateImportQuery = `
UPDATE imports SET spec=$2, status=$3 WHERE id=$1 RETURNING 1;`
