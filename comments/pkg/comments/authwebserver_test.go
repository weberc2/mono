package comments

import (
	"errors"
	"net/http"
	"testing"

	pz "github.com/weberc2/httpeasy"
	"github.com/weberc2/mono/comments/pkg/auth/client"
	"github.com/weberc2/mono/comments/pkg/comments/testsupport"
)

func TestAuthWebServer(t *testing.T) {
	for _, testCase := range []struct {
		name     string
		method   func(*AuthWebServer) pz.Route
		optional bool
	}{
		{
			name:     "replies",
			method:   (*AuthWebServer).RepliesRoute,
			optional: true,
		},
		{
			name:     "delete-confirm",
			method:   (*AuthWebServer).DeleteConfirmRoute,
			optional: false,
		},
		{
			name:     "delete",
			method:   (*AuthWebServer).DeleteRoute,
			optional: false,
		},
		{
			name:     "reply-form",
			method:   (*AuthWebServer).ReplyFormRoute,
			optional: false,
		},
		{
			name:     "reply",
			method:   (*AuthWebServer).ReplyRoute,
			optional: false,
		},
		{
			name:     "edit-form",
			method:   (*AuthWebServer).EditFormRoute,
			optional: false,
		},
		{
			name:     "edit",
			method:   (*AuthWebServer).EditRoute,
			optional: false,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			rsp := testCase.method(&AuthWebServer{
				WebServer: WebServer{
					Comments: CommentsModel{
						CommentsStore: testsupport.CommentsStoreFake{},
					},
				},
				Auth: Auth{
					AuthType: client.ConstantAuthType(client.ResultErr(
						"ERR",
						errors.New("ERR"),
					)),
				},
			}).Handler(pz.Request{Headers: make(http.Header)})

			if testCase.optional && rsp.Status == 401 {
				t.Fatal("expected optional authentication, but got `401`")
			}
			if !testCase.optional && rsp.Status != 401 {
				t.Fatalf("auth required--wanted `401`; found `%d`", rsp.Status)
			}
		})
	}
}
