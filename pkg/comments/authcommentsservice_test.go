package comments

import (
	"errors"
	"net/http"
	"testing"
	"time"

	pz "github.com/weberc2/httpeasy"
	"github.com/weberc2/mono/pkg/auth/client"
	"github.com/weberc2/mono/pkg/comments/testsupport"
)

func TestAuthCommentsService(t *testing.T) {
	for _, testCase := range []struct {
		name         string
		method       func(*AuthCommentsService) pz.Route
		requiresAuth bool
	}{
		{
			name:         "replies",
			method:       (*AuthCommentsService).RepliesRoute,
			requiresAuth: false,
		},
		{
			name:         "get",
			method:       (*AuthCommentsService).GetRoute,
			requiresAuth: false,
		},
		{
			name:         "put",
			method:       (*AuthCommentsService).PutRoute,
			requiresAuth: true,
		},
		{
			name:         "delete",
			method:       (*AuthCommentsService).DeleteRoute,
			requiresAuth: true,
		},
		{
			name:         "update",
			method:       (*AuthCommentsService).UpdateRoute,
			requiresAuth: true,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			rsp := testCase.method(&AuthCommentsService{
				CommentsService: CommentsService{
					Comments: CommentsModel{
						CommentsStore: testsupport.CommentsStoreFake{},
						TimeFunc:      func() time.Time { return now },
					},
					TimeFunc: func() time.Time { return now },
				},
				Auth: Auth{
					AuthType: client.ConstantAuthType(client.ResultErr(
						"ERR",
						errors.New("ERR"),
					)),
				},
			}).Handler(pz.Request{Headers: make(http.Header)})

			if !testCase.requiresAuth && rsp.Status == 401 {
				t.Fatalf("expected no auth, but got `401`")
			}

			if testCase.requiresAuth && rsp.Status != 401 {
				t.Fatalf("auth required--wanted `401`; found `%d`", rsp.Status)
			}
		})
	}
}
