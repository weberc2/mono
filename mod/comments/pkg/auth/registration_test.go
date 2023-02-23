package auth

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	pz "github.com/weberc2/httpeasy"
	"github.com/weberc2/mono/mod/comments/pkg/auth/testsupport"
	"github.com/weberc2/mono/mod/comments/pkg/auth/types"
	. "github.com/weberc2/mono/mod/comments/pkg/prelude"
)

func TestWebServer_Registration(t *testing.T) {
	for _, testCase := range []*routeTestCase{
		{
			name: "main-form",
			route: func(ws *WebServer) pz.Handler {
				return ws.RegistrationFormRoute().Handler
			},
			wantedResponse: response{
				status: http.StatusOK,
				body: wantedFormDocument(&formDocument{
					title:  "Registration",
					action: "/registration",
					fields: []input{{
						name:  "username",
						type_: "text",
						label: "Username",
					}, {
						name:  "email",
						type_: "text",
						label: "Email",
					}},
				}),
			},
		},
		{
			name: "main-handler",
			route: func(ws *WebServer) pz.Handler {
				return ws.RegistrationHandlerRoute().Handler
			},
			request: pz.Request{
				Body: strings.NewReader(encodeForm(
					kv{"username", string(testsupport.User)},
					kv{"email", testsupport.Email},
				)),
			},
			wantedResponse: response{
				status: http.StatusAccepted,
				body: testsupport.WantedString(Must(ackPage(
					"Registration",
				))),
			},
			wantedNotifications: []*types.Notification{
				&testsupport.RegistrationNotification,
			},
		},
		{
			name: "main-handler-exists",
			route: func(ws *WebServer) pz.Handler {
				return ws.RegistrationHandlerRoute().Handler
			},
			users: testsupport.UserStoreFake{
				testsupport.User: &types.UserEntry{
					User:         testsupport.User,
					Email:        testsupport.Email,
					PasswordHash: testsupport.GoodPasswordHash,
				},
			},
			request: pz.Request{
				Body: strings.NewReader(encodeForm(
					kv{"username", string(testsupport.User)},
					kv{"email", testsupport.Email},
				)),
			},
			wantedResponse: response{
				status: http.StatusConflict,
				body: wantedFormDocument(&formDocument{
					title:        "Registration",
					action:       "/registration",
					errorMessage: "user already exists",
					fields: []input{{
						name:  "username",
						type_: "text",
						label: "Username",
					}, {
						name:  "email",
						type_: "text",
						label: "Email",
					}},
				}),
			},
			wantedUsers: []types.Credentials{{
				User:     testsupport.User,
				Email:    testsupport.Email,
				Password: testsupport.GoodPassword,
			}},
			wantedNotifications: []*types.Notification{},
		},
		{
			name: "confirmation-form",
			request: pz.Request{
				URL: Must(url.Parse("https://auth.example.org?t=<token>")),
			},
			route: func(ws *WebServer) pz.Handler {
				return ws.RegistrationConfirmationFormRoute().Handler
			},
			wantedResponse: response{
				status: http.StatusOK,
				body: wantedFormDocument(&formDocument{
					title:  "Confirm Registration",
					action: "/registration/confirm",
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
			route: func(ws *WebServer) pz.Handler {
				return ws.RegistrationConfirmationHandlerRoute().Handler
			},
			request: pz.Request{
				URL: Must(url.Parse(
					"https://auth.example.org/registration/confirm",
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
				User:     testsupport.User,
				Email:    testsupport.Email,
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
