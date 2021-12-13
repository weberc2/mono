package comments

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/weberc2/comments/pkg/testsupport"
	"github.com/weberc2/comments/pkg/types"
)

const (
	goodBody = "sufficiently long body"
)

func TestCommentsModel_Delete(t *testing.T) {
	for _, testCase := range []struct {
		name        string
		state       testsupport.CommentsStoreFake
		post        types.PostID
		comment     types.CommentID
		wantedState testsupport.CommentsStoreFake
		wantedErr   types.WantedError
	}{
		{
			name: "simple",
			state: testsupport.CommentsStoreFake{
				"post": {
					"id": {
						ID:       "id",
						Post:     "post",
						Parent:   "parent",
						Author:   "author",
						Created:  someTime,
						Modified: someTime,
						Deleted:  false,
					},
				},
			},
			post:    "post",
			comment: "id",
			wantedState: testsupport.CommentsStoreFake{
				"post": {
					"id": {
						ID:       "id",
						Post:     "post",
						Parent:   "parent",
						Author:   "author",
						Created:  someTime,
						Modified: now,
						Deleted:  true,
					},
				},
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			model := CommentsModel{
				CommentsStore: testCase.state,
				IDFunc:        func() types.CommentID { return "comment" },
				TimeFunc:      func() time.Time { return now },
			}

			if testCase.wantedErr == nil {
				testCase.wantedErr = types.NilError{}
			}
			if err := testCase.wantedErr.CompareErr(
				model.Delete(testCase.post, testCase.comment),
			); err != nil {
				t.Fatal(err)
			}

			t.Logf("Wanted state: %s", jsonify(testCase.wantedState.List()))
			t.Logf("Actual state: %s", jsonify(testCase.state.List()))
			t.Fail()

			if err := testCase.wantedState.Compare(
				testCase.state,
			); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func jsonify(x interface{}) []byte {
	data, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		panic(err)
	}
	return data
}

func TestCommentsModel_Put(t *testing.T) {
	for _, testCase := range []struct {
		name          string
		input         types.Comment
		wantedComment *types.Comment
		wantedErr     types.WantedError
	}{
		{
			name: "base case",
			input: types.Comment{
				Post:   "post",
				Author: "user",
				Body:   goodBody,
			},
			wantedComment: &types.Comment{
				Post:     "post",
				ID:       "comment",
				Author:   "user",
				Created:  now,
				Modified: now,
				Body:     goodBody,
			},
		},
		{
			name:      "missing `post` field",
			input:     types.Comment{Author: "user", Body: goodBody},
			wantedErr: ErrInvalidPost,
		},
		{
			name: "ignores `ID`, `Created`, `Modified`, and `Deleted` fields",
			input: types.Comment{
				Post:     "post",
				ID:       "evil-id",
				Author:   "user",
				Created:  someTime,
				Modified: someTime,
				Deleted:  true,
				Body:     goodBody,
			},
			wantedComment: &types.Comment{
				Post:     "post",
				ID:       "comment",
				Author:   "user",
				Created:  now,
				Modified: now,
				Deleted:  false,
				Body:     goodBody,
			},
		},
		{
			name:      "body too short",
			input:     types.Comment{Post: "post", Author: "author", Body: ""},
			wantedErr: ErrBodyTooShort,
		},
		{
			name: "body too long",
			input: types.Comment{
				Post:   "post",
				Author: "author",
				Body:   strings.Repeat("x", bodySizeMax+1),
			},
			wantedErr: ErrBodyTooLong,
		},
		{
			name: "html escape body",
			input: types.Comment{
				Post:   "post",
				Author: "user",
				Body:   "<script></script>",
			},
			wantedComment: &types.Comment{
				Post:     "post",
				ID:       "comment",
				Author:   "user",
				Created:  now,
				Modified: now,
				Body:     "&lt;script&gt;&lt;/script&gt;",
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			store := testsupport.CommentsStoreFake{}
			model := CommentsModel{
				CommentsStore: store,
				IDFunc:        func() types.CommentID { return "comment" },
				TimeFunc:      func() time.Time { return now },
			}

			c, err := model.Put(&testCase.input)

			if testCase.wantedErr == nil ||
				testCase.wantedErr == (types.NilError{}) {
				if err := (types.NilError{}).CompareErr(err); err != nil {
					t.Fatal(err)
				}
				if err := testCase.wantedComment.Compare(c); err != nil {
					t.Fatal(err)
				}
				if err := store.Contains(testCase.wantedComment); err != nil {
					t.Fatal(err)
				}
			} else if err := testCase.wantedErr.CompareErr(err); err != nil {
				t.Fatal(err)
			}
		})
	}
}

var (
	someTime = time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	now      = time.Date(2022, 12, 31, 0, 0, 0, 0, time.UTC)
)
