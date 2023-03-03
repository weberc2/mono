package auth

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	pz "github.com/weberc2/httpeasy"
	"github.com/weberc2/mono/comments/pkg/auth/testsupport"
	"github.com/weberc2/mono/comments/pkg/auth/types"
	. "github.com/weberc2/mono/comments/pkg/prelude"
)

func TestWebServer_PasswordReset(t *testing.T) {
	for _, testCase := range []*routeTestCase{
		{
			name: "main-form",
			route: func(ws *WebServer) pz.Handler {
				return ws.PasswordResetFormRoute().Handler
			},
			wantedResponse: response{
				status: http.StatusOK,
				body: wantedFormDocument(&formDocument{
					title:  "Password Reset",
					action: "/password-reset",
					fields: []input{{
						name:  "username",
						type_: "text",
						label: "Username",
					}},
				}),
			},
		},
		{
			name: "main-handler",
			route: func(ws *WebServer) pz.Handler {
				return ws.PasswordResetHandlerRoute().Handler
			},
			users: testsupport.UserStoreFake{
				testsupport.User: &types.UserEntry{
					User:         testsupport.User,
					Email:        testsupport.Email,
					PasswordHash: testsupport.GoodPasswordHash,
				},
			},
			request: pz.Request{
				Body: strings.NewReader(encodeForm(kv{
					"username",
					string(testsupport.User),
				})),
			},
			wantedResponse: response{
				status: http.StatusAccepted,
				body: testsupport.WantedString(Must(ackPage(
					"Password Reset",
				))),
			},
			wantedUsers: []types.Credentials{{
				User:     testsupport.User,
				Email:    testsupport.Email,
				Password: testsupport.GoodPassword,
			}},
			wantedNotifications: []*types.Notification{
				&testsupport.PasswordResetNotification,
			},
		},
		{
			name: "main-handler-not-found",
			route: func(ws *WebServer) pz.Handler {
				return ws.PasswordResetHandlerRoute().Handler
			},
			request: pz.Request{
				Body: strings.NewReader(encodeForm(kv{
					"username",
					string(testsupport.User),
				})),
			},

			// if we're trying to reset the password for a user which doesn't
			// exist, then we want to send the same response pack to the user
			// that we would have sent if the user did exist, but we won't
			// send a notification.
			wantedResponse: response{
				status: http.StatusAccepted,
				body: testsupport.WantedString(Must(ackPage(
					"Password Reset",
				))),
			},
			wantedNotifications: []*types.Notification{},
		},
		{
			name: "confirmation-form",
			request: pz.Request{
				URL: Must(url.Parse("https://auth.example.org?t=<token>")),
			},
			route: func(ws *WebServer) pz.Handler {
				return ws.PasswordResetConfirmationFormRoute().Handler
			},
			wantedResponse: response{
				status: http.StatusOK,
				body: wantedFormDocument(&formDocument{
					title:  "Confirm Password Reset",
					action: "/password-reset/confirm",
					fields: []input{{
						name:  "password",
						type_: "password",
						label: "Password",
					}, {
						name:  "token",
						type_: "hidden",
						value: "<token>",
					}},
				}),
			},
		},
		{
			name: "confirmation-handler",
			users: testsupport.UserStoreFake{
				"user": &types.UserEntry{
					User:  "user",
					Email: "user@example.org",
				},
			},
			route: func(ws *WebServer) pz.Handler {
				return ws.PasswordResetConfirmationHandlerRoute().Handler
			},
			request: pz.Request{
				URL: Must(url.Parse(
					"https://auth.example.org/password-reset/confirm",
				)),
				Body: strings.NewReader(encodeForm(
					kv{"password", testsupport.GoodPassword},
					kv{"token", testsupport.ResetToken},
				)),
			},
			wantedResponse: response{
				status: http.StatusSeeOther,
				body:   testsupport.WantedString("303 See Other"),
				headers: http.Header{
					"Location": []string{testsupport.DefaultRedirectLocation},
				},
			},
			wantedUsers: []types.Credentials{{
				User:     "user",
				Email:    "user@example.org",
				Password: testsupport.GoodPassword,
			}},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if err := testCase.run(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
