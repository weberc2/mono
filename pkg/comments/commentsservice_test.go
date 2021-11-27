package comments

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/weberc2/comments/pkg/testsupport"
	"github.com/weberc2/comments/pkg/types"
	pz "github.com/weberc2/httpeasy"
)

func TestPutComment(t *testing.T) {
	now := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, testCase := range []struct {
		name         string
		input        string
		wantedStatus int
		wantedBody   WantedData
	}{
		{
			name:         "creates comment",
			input:        `{"body": "great comment"}`,
			wantedStatus: http.StatusCreated,
			wantedBody: &types.Comment{
				ID:       "comment",
				Author:   "user",
				Parent:   "",
				Created:  now,
				Modified: now,
				Body:     "great comment",
			},
		},
		{
			// If the caller tries to pass explicit values for id, author,
			// created, or modified fields, these fields are ignored in favor
			// of authoritative sources (`id` is generated by the underlying
			// CommentStore, `author` defers to the `User` header, and
			// `created` and `modified` are set to the current time).
			name: "ignores id, author, created, modified",
			input: `{
	"id": "asdf",
	"author": "foo",
	"body": "great comment",
	"created": "1970-01-01T00:00:00.000000Z",
	"modified": "1970-01-01T00:00:00.000000Z"
}`,
			wantedStatus: http.StatusCreated,
			wantedBody: &types.Comment{
				ID:       "comment",
				Author:   "user",
				Body:     "great comment",
				Parent:   "",
				Created:  now,
				Modified: now,
			},
		},
		{
			name:         "errors serialized correctly",
			input:        `{"body": ""}`,
			wantedStatus: 400,
			wantedBody:   ErrBodyTooShort,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			commentsService := CommentsService{
				Comments: CommentsModel{
					CommentsStore: &ObjectCommentsStore{
						ObjectStore: testsupport.ObjectStoreFake{},
						Bucket:      "bucket",
						Prefix:      "prefix",
					},
					TimeFunc: func() time.Time { return now },
					IDFunc:   func() types.CommentID { return "comment" },
				},
				TimeFunc: func() time.Time { return now },
			}

			rsp := commentsService.PutComment(pz.Request{
				Vars:    map[string]string{"post-id": "post"},
				Headers: http.Header{"User": []string{"user"}},
				Body:    strings.NewReader(testCase.input),
			})

			if rsp.Status != testCase.wantedStatus {
				data, err := readAll(rsp.Data)
				if err != nil {
					t.Logf("error reading response body: %v", err)
				}
				for _, logging := range rsp.Logging {
					data, err := json.Marshal(logging)
					if err != nil {
						t.Logf("error marshalling logging as JSON: %v", err)
					}
					t.Logf("log: %s", data)
				}
				t.Fatalf(
					"Response.Status: wanted `%d`; found `%d`: %s",
					testCase.wantedStatus,
					rsp.Status,
					data,
				)
			}

			data, err := readAll(rsp.Data)
			if err != nil {
				t.Fatalf("Response.Data: reading serializer: %v", err)
			}

			if err := testCase.wantedBody.CompareData(data); err != nil {
				t.Errorf("Response.Data: %v", err)
				t.Logf("Actual body: %s", data)
			}
		})
	}
}

func readAll(s pz.Serializer) ([]byte, error) {
	writerTo, err := s()
	if err != nil {
		return nil, fmt.Errorf("serializing: %w", err)
	}

	var b bytes.Buffer
	if _, err := writerTo.WriteTo(&b); err != nil {
		return nil, fmt.Errorf("copying data to buffer: %w", err)
	}

	return b.Bytes(), nil
}

type WantedData interface {
	CompareData([]byte) error
}
