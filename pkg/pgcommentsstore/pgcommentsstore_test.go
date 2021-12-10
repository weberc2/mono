package pgcommentsstore

import (
	"fmt"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/weberc2/comments/pkg/types"
)

func TestPut(t *testing.T) {
	store, err := testPGCommentsStore()
	if err != nil {
		t.Fatal(err)
	}

	for _, testCase := range []struct {
		name          string
		state         []*types.Comment
		input         types.Comment
		wantedComment *types.Comment
		wantedError   types.WantedError
	}{
		{
			name: "simple",
			input: types.Comment{
				ID:       "id",
				Post:     "post",
				Parent:   "",
				Author:   "author",
				Created:  someDate,
				Modified: someDate,
				Body:     "body",
			},
			wantedComment: &types.Comment{
				ID:       "id",
				Post:     "post",
				Parent:   "",
				Author:   "author",
				Created:  someDate,
				Modified: someDate,
				Body:     "body",
			},
			wantedError: nil,
		},
		{
			// GIVEN a comment with ID `id` exists
			// WHEN we try to Put() a comment with the same ID
			// THEN expect we get a `CommentExistsErr`
			name: "unique ids",
			state: []*types.Comment{
				{
					ID:       "id",
					Post:     "post",
					Parent:   "",
					Author:   "author",
					Created:  someDate,
					Modified: someDate,
					Body:     "body",
				},
			},
			input: types.Comment{
				ID:       "id",
				Post:     "post",
				Parent:   "",
				Author:   "author",
				Created:  someDate,
				Modified: someDate,
				Body:     "body",
			},
			wantedComment: nil,
			wantedError: &types.CommentExistsErr{
				Post:    "post",
				Comment: "id",
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if err := resetTable(store); err != nil {
				t.Fatal(err)
			}

			for _, comment := range testCase.state {
				if _, err := store.Put(comment); err != nil {
					t.Fatalf(
						"unexpected error preparing test database state: %v",
						err,
					)
				}
			}

			c, err := store.Put(&testCase.input)

			if err := testCase.wantedComment.Compare(c); err != nil {
				t.Fatal(err)
			}

			if testCase.wantedError == nil {
				testCase.wantedError = types.NilError{}
			}
			if err := testCase.wantedError.CompareErr(err); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestComment(t *testing.T) {
	store, err := testPGCommentsStore()
	if err != nil {
		t.Fatal(err)
	}

	input := types.Comment{
		ID:       "id",
		Post:     "post",
		Parent:   "",
		Author:   "author",
		Created:  someDate,
		Modified: someDate,
		Body:     "body",
	}

	if _, err := store.Put(&input); err != nil {
		t.Fatalf("unexpected error putting comment: %v", err)
	}

	found, err := store.Comment(input.Post, input.ID)
	if err != nil {
		t.Fatalf("unexpected error fetching comment: %v", err)
	}

	if err := input.Compare(found); err != nil {
		t.Fatal(err)
	}
}

func TestDelete(t *testing.T) {
	store, err := testPGCommentsStore()
	if err != nil {
		t.Fatal(err)
	}

	input := types.Comment{
		ID:       "id",
		Post:     "post",
		Parent:   "",
		Author:   "author",
		Created:  someDate,
		Modified: someDate,
		Body:     "body",
	}

	if _, err := store.Put(&input); err != nil {
		t.Fatalf("unexpected error putting comment: %v", err)
	}

	if err := store.Delete(input.Post, input.ID); err != nil {
		t.Fatalf("unexpected error deleting comment: %v", err)
	}

	_, err = store.Comment(input.Post, input.ID)
	wanted := types.CommentNotFoundErr{Post: "post", Comment: "id"}
	if err := wanted.CompareErr(err); err != nil {
		t.Fatal(err)
	}
}

func TestReplies(t *testing.T) {
	store, err := testPGCommentsStore()
	if err != nil {
		t.Fatal(err)
	}
	for _, testCase := range []struct {
		name          string
		state         []*types.Comment
		post          types.PostID
		parent        types.CommentID
		wantedReplies []*types.Comment
		wantedErr     types.WantedError
	}{
		{
			name:          "empty",
			state:         nil,
			post:          "post",
			parent:        "parent",
			wantedReplies: nil,
			wantedErr:     nil,
		},
		{
			name: "toplevel-simple",
			state: []*types.Comment{
				{
					ID:       "0",
					Post:     "post",
					Parent:   "",
					Author:   "author",
					Created:  someDate,
					Modified: someDate,
					Body:     "body",
				},
				{
					ID:       "1",
					Post:     "post",
					Parent:   "",
					Author:   "author",
					Created:  someDate,
					Modified: someDate,
					Body:     "body",
				},
				{
					ID:       "different-post-comment",
					Post:     "some-other-post",
					Parent:   "",
					Author:   "author",
					Created:  someDate,
					Modified: someDate,
					Body:     "body",
				},
			},
			post:   "post",
			parent: "",
			wantedReplies: []*types.Comment{
				{
					ID:       "0",
					Post:     "post",
					Parent:   "",
					Author:   "author",
					Created:  someDate,
					Modified: someDate,
					Body:     "body",
				},
				{
					ID:       "1",
					Post:     "post",
					Parent:   "",
					Author:   "author",
					Created:  someDate,
					Modified: someDate,
					Body:     "body",
				},
			},
			wantedErr: nil,
		},
		{
			// GIVEN two posts in the database, one with the parent `parent`
			//   and the other with no parent (i.e., "toplevel")
			// WHEN replies are fetched for the comment `parent`
			// THEN expect only the comment whose parent is `parent` is
			//   returned.
			name: "with-parent",
			state: []*types.Comment{
				{
					ID:       "parent",
					Post:     "post",
					Parent:   "",
					Author:   "author",
					Created:  someDate,
					Modified: someDate,
					Body:     "body",
				},
				{
					ID:       "child",
					Post:     "post",
					Parent:   "parent",
					Author:   "author",
					Created:  someDate,
					Modified: someDate,
					Body:     "body",
				},
				{
					ID:       "unrelated",
					Post:     "post",
					Parent:   "",
					Author:   "author",
					Created:  someDate,
					Modified: someDate,
					Body:     "body",
				},
				{
					ID:       "different-post-comment",
					Post:     "some-other-post",
					Parent:   "",
					Author:   "author",
					Created:  someDate,
					Modified: someDate,
					Body:     "body",
				},
			},
			post:   "post",
			parent: "parent",
			wantedReplies: []*types.Comment{
				{
					ID:       "child",
					Post:     "post",
					Parent:   "parent",
					Author:   "author",
					Created:  someDate,
					Modified: someDate,
					Body:     "body",
				},
			},
			wantedErr: nil,
		},
		{
			name: "recursive",
			state: []*types.Comment{
				{
					ID:       "parent",
					Post:     "post",
					Parent:   "",
					Author:   "author",
					Created:  someDate,
					Modified: someDate,
					Body:     "body",
				},
				{
					ID:       "child",
					Post:     "post",
					Parent:   "parent",
					Author:   "author",
					Created:  someDate,
					Modified: someDate,
					Body:     "body",
				},
				{
					ID:       "grandchild",
					Post:     "post",
					Parent:   "child",
					Author:   "author",
					Created:  someDate,
					Modified: someDate,
					Body:     "body",
				},
				{
					ID:       "unrelated",
					Post:     "post",
					Parent:   "",
					Author:   "author",
					Created:  someDate,
					Modified: someDate,
					Body:     "body",
				},
				{
					ID:       "different-post-comment",
					Post:     "some-other-post",
					Parent:   "",
					Author:   "author",
					Created:  someDate,
					Modified: someDate,
					Body:     "body",
				},
			},
			post:   "post",
			parent: "parent",
			wantedReplies: []*types.Comment{
				{
					ID:       "child",
					Post:     "post",
					Parent:   "parent",
					Author:   "author",
					Created:  someDate,
					Modified: someDate,
					Body:     "body",
				},
				{
					ID:       "grandchild",
					Post:     "post",
					Parent:   "child",
					Author:   "author",
					Created:  someDate,
					Modified: someDate,
					Body:     "body",
				},
			},
			wantedErr: nil,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if err := resetTable(store); err != nil {
				t.Fatal(err)
			}

			for _, comment := range testCase.state {
				if _, err := store.Put(comment); err != nil {
					t.Fatalf(
						"unexpected error preparing test database state: %v",
						err,
					)
				}
			}

			found, err := store.Replies(testCase.post, testCase.parent)

			if testCase.wantedErr == nil {
				testCase.wantedErr = types.NilError{}
			}
			if err := testCase.wantedErr.CompareErr(err); err != nil {
				t.Fatalf("comparing errors: %v", err)
			}

			if err := types.CompareComments(
				testCase.wantedReplies,
				found,
			); err != nil {
				t.Fatalf("comparing replies: %v", err)
			}
		})
	}
}

func testPGCommentsStore() (*PGCommentsStore, error) {
	pgcs, err := OpenEnv()
	if err != nil {
		return nil, fmt.Errorf("opening test database: %w", err)
	}
	return pgcs, resetTable(pgcs)
}

func resetTable(store *PGCommentsStore) error {
	if err := store.DropTable(); err != nil {
		return err
	}
	return store.EnsureTable()
}

var someDate = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
