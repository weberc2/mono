package pgtokenstore

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/weberc2/auth/pkg/types"
)

type PGTokenStore sql.DB

func OpenEnv() (*PGTokenStore, error) {
	db, err := sql.Open(
		"postgres",
		fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			getEnv("PG_HOST", "localhost"),
			getEnv("PG_PORT", "5432"),
			getEnv("PG_USER", "postgres"),
			getEnv("PG_PASS", ""),
			getEnv("PG_DB_NAME", "postgres"),
			getEnv("PG_SSL_MODE", "disable"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("opening postgres database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("pinging postgres database: %w", err)
	}

	return (*PGTokenStore)(db), nil
}

func getEnv(env, def string) string {
	x := os.Getenv(env)
	if x == "" {
		return def
	}
	return x
}

func (pgts *PGTokenStore) EnsureTable() error {
	if _, err := (*sql.DB)(pgts).Exec(
		"CREATE TABLE IF NOT EXISTS tokens (" +
			"token VARCHAR(9000) NOT NULL PRIMARY KEY, " +
			"expires TIMESTAMP NOT NULL)",
	); err != nil {
		return fmt.Errorf("creating `tokens` postgres table: %w", err)
	}
	return nil
}

func (pgts *PGTokenStore) DropTable() error {
	if _, err := (*sql.DB)(pgts).Exec(
		"DROP TABLE IF EXISTS tokens",
	); err != nil {
		return fmt.Errorf("dropping table `tokens`: %w", err)
	}
	return nil
}

func (pgts *PGTokenStore) ClearTable() error {
	if _, err := (*sql.DB)(pgts).Exec("DELETE FROM tokens"); err != nil {
		return fmt.Errorf("clearing `tokens` postgres table: %w", err)
	}
	return nil
}

func (pgts *PGTokenStore) ResetTable() error {
	if err := pgts.DropTable(); err != nil {
		return err
	}
	return pgts.EnsureTable()
}

func (pgts *PGTokenStore) Put(token string, expires time.Time) error {
	if _, err := (*sql.DB)(pgts).Exec(
		"INSERT INTO tokens (token, expires) VALUES($1, $2)",
		token,
		expires.Format(time.RFC3339),
	); err != nil {
		const errUniqueViolation = "23505"
		if err, ok := err.(*pq.Error); ok && err.Code == errUniqueViolation {
			return types.ErrTokenExists
		}
		return fmt.Errorf("inserting token into postgres: %w", err)
	}
	return nil
}

func (pgts *PGTokenStore) Exists(token string) error {
	var dummy string
	if err := (*sql.DB)(pgts).QueryRow(
		"SELECT true FROM tokens WHERE token = $1",
		token,
	).Scan(&dummy); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = types.ErrTokenNotFound
		}
		return fmt.Errorf("checking for token in postgres: %w", err)
	}
	return nil
}

func (pgts *PGTokenStore) Delete(token string) error {
	if err := (*sql.DB)(pgts).QueryRow(
		"DELETE FROM tokens WHERE token = $1 RETURNING token",
		token,
	).Scan(&token); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = types.ErrTokenNotFound
		}
		return fmt.Errorf("deleting token from postgres: %w", err)
	}
	return nil
}

// DeleteExpired deletes all tokens that expired before `now`.
func (pgts *PGTokenStore) DeleteExpired(now time.Time) error {
	if _, err := (*sql.DB)(pgts).Exec(
		"DELETE FROM tokens WHERE expires < $1",
		now,
	); err != nil {
		return fmt.Errorf("deleting expired tokens from postgres: %w", err)
	}
	return nil
}

func (pgts *PGTokenStore) List() ([]types.Token, error) {
	// we don't want to return a `nil` slice because that gets JSON-marshaled
	// to `null` instead of `[]`.
	entries := []types.Token{}

	rows, err := (*sql.DB)(pgts).Query("SELECT token, expires FROM tokens")
	if err != nil {
		return nil, fmt.Errorf("querying tokens from postgres: %w", err)
	}

	for rows.Next() {
		var entry types.Token
		var expiresString string
		if err := rows.Scan(&entry.Token, &expiresString); err != nil {
			return nil, fmt.Errorf("querying tokens from postgres: %w", err)
		}
		exp, err := time.Parse(time.RFC3339, expiresString)
		if err != nil {
			return nil, fmt.Errorf(
				"querying tokens from postgres: parsing `expires` field: %w",
				err,
			)
		}
		entry.Expires = exp
		entries = append(entries, entry)
	}

	return entries, err
}

var (
	_ types.TokenStore = &PGTokenStore{}
)
