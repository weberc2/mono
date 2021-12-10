package pgcommentsstore

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/lib/pq"
	"github.com/weberc2/comments/pkg/types"
)

const errUniqueViolation = "23505"

type PGCommentsStore struct {
	DB *sql.DB
}

func OpenEnv() (*PGCommentsStore, error) {
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

	return &PGCommentsStore{db}, nil
}

func getEnv(env, def string) string {
	x := os.Getenv(env)
	if x == "" {
		return def
	}
	return x
}

func (pgcs *PGCommentsStore) DropTable() error {
	if _, err := pgcs.DB.Exec("DROP TABLE IF EXISTS comments"); err != nil {
		return fmt.Errorf("dropping table `comments`: %w", err)
	}
	return nil
}

func (pgcs *PGCommentsStore) EnsureTable() error {
	if _, err := pgcs.DB.Exec(
		"CREATE TABLE IF NOT EXISTS comments (" +
			"id VARCHAR(255) NOT NULL PRIMARY KEY, " +
			"post TEXT NOT NULL, " +
			"parent VARCHAR(255) NOT NULL, " +
			"author VARCHAR(255) NOT NULL, " +
			"created VARCHAR(255) NOT NULL, " +
			"modified VARCHAR(255) NOT NULL, " +
			"body TEXT NOT NULL)",
	); err != nil {
		return fmt.Errorf("creating `comments` postgres table: %w", err)
	}
	return nil
}

func (pgcs *PGCommentsStore) ClearTable() error {
	if _, err := pgcs.DB.Exec("DELETE FROM comments"); err != nil {
		return fmt.Errorf("clearing `comments` postgres table: %w", err)
	}
	return nil
}

func (pgcs *PGCommentsStore) ResetTable() error {
	if err := pgcs.DropTable(); err != nil {
		return err
	}
	return pgcs.EnsureTable()
}

func (pgcs *PGCommentsStore) Put(c *types.Comment) (*types.Comment, error) {
	if _, err := pgcs.DB.Exec(
		"INSERT INTO comments "+
			"(id, post, parent, author, created, modified, body) VALUES"+
			"($1, $2, $3, $4, $5, $6, $7);",
		c.ID,
		c.Post,
		c.Parent,
		c.Author,
		c.Created.Format(time.RFC3339),
		c.Modified.Format(time.RFC3339),
		c.Body,
	); err != nil {
		if err, ok := err.(*pq.Error); ok && err.Code == errUniqueViolation {
			return nil, &types.CommentExistsErr{Post: c.Post, Comment: c.ID}
		}
		return nil, fmt.Errorf(
			"inserting comment into postgres: %w",
			err,
		)
	}

	return c, nil
}

func (pgcs *PGCommentsStore) Comment(
	p types.PostID,
	c types.CommentID,
) (*types.Comment, error) {
	var comment types.Comment
	if err := scanComment(&comment, pgcs.DB.QueryRow(
		"SELECT id, post, parent, author, created, modified, body "+
			"FROM comments WHERE post = $1 AND id = $2",
		p,
		c,
	)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &types.CommentNotFoundErr{Post: p, Comment: c}
		}
		return nil, fmt.Errorf("fetching comment from postgres: %w", err)
	}

	return &comment, nil
}

func (pgcs *PGCommentsStore) Replies(
	p types.PostID,
	parent types.CommentID,
) ([]*types.Comment, error) {
	rows, err := pgcs.DB.Query(
		`WITH RECURSIVE t AS (
	SELECT * FROM comments WHERE post = $1 AND parent = $2 UNION
	SELECT comments.* FROM comments JOIN t ON
	comments.post = t.post AND comments.parent = t.id
) SELECT * FROM t`,
		p,
		parent,
	)
	if err != nil {
		return nil, fmt.Errorf("querying replies from postgres: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("PGCommentsStore.Replies(): closing sql.Rows: %v", err)
		}
	}()

	// we want to initialize these so that an empty slice serializes to `[]`
	// instead of `null`.
	buf := []types.Comment{}  // put all results in a single allocation
	out := []*types.Comment{} // every item in `out` points into `buf`
	for i := 0; rows.Next(); i++ {
		buf = append(buf, types.Comment{})
		if err := scanComment(&buf[i], rows); err != nil {
			return nil, fmt.Errorf(
				"scanning postgres row into comment: %w",
				err,
			)
		}
		out = append(out, &buf[i])
	}
	return out, nil
}

func (pgcs *PGCommentsStore) Delete(p types.PostID, c types.CommentID) error {
	if _, err := pgcs.DB.Exec(
		"DELETE FROM comments WHERE post = $1 AND id = $2",
		p,
		c,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &types.CommentNotFoundErr{Post: p, Comment: c}
		}
		return fmt.Errorf("deleting comment from postgres: %w", err)
	}
	return nil
}

func scanComment(
	c *types.Comment,
	s interface{ Scan(...interface{}) error },
) error {
	var createdString, modifiedString string
	if err := s.Scan(
		&c.ID,
		&c.Post,
		&c.Parent,
		&c.Author,
		&createdString,
		&modifiedString,
		&c.Body,
	); err != nil {
		return err
	}
	created, err := time.Parse(time.RFC3339, createdString)
	if err != nil {
		return fmt.Errorf(
			"parsing `created` time from `%s`: %v",
			createdString,
			err,
		)
	}
	modified, err := time.Parse(time.RFC3339, modifiedString)
	if err != nil {
		return fmt.Errorf(
			"parsing `modified` time from `%s`: %v",
			modifiedString,
			err,
		)
	}
	c.Created = created
	c.Modified = modified
	return nil
}
