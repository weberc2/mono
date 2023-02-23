package auth

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	pz "github.com/weberc2/httpeasy"
	"github.com/weberc2/mono/mod/auth/pkg/auth/testsupport"
	"github.com/weberc2/mono/mod/auth/pkg/auth/types"
	. "github.com/weberc2/mono/mod/auth/pkg/prelude"
)

func TestConfirmationFlow(t *testing.T) {
	for _, testCase := range []struct {
		name         string
		flowParams   confirmationFlowParams
		users        testsupport.UserStoreFake
		request      pz.Request
		route        func(*WebServer, *confirmationFlow) pz.Route
		wantedUsers  []types.Credentials
		wantedStatus int
		wantedBody   testsupport.WantedBody
	}{
		{
			name:    "main-form",
			request: pz.Request{URL: Must(url.Parse("https://example.org"))},
			route:   mainFormRoute,
			flowParams: confirmationFlowParams{
				activity: "Activity",
				basePath: "/activity",
				fields: []field{{
					ID:    "username",
					Label: "Username",
				}, {
					ID:    "password",
					Label: "Password",
					Type:  "password",
				}, {
					ID:     "token",
					Hidden: true,
					Value:  "<token>",
				}},
			},
			wantedStatus: http.StatusOK,
			wantedBody: wantedFormDocument(&formDocument{
				title:  "Activity",
				action: "/activity",
				fields: []input{{
					name:  "username",
					type_: "text",
					value: "",
					label: "Username",
				}, {
					name:  "password",
					type_: "password",
					value: "",
					label: "Password",
				}, {
					name:  "token",
					type_: "hidden",
					value: "<token>",
					label: "",
				}},
			}),
		},
		{
			name: "main-handler-simple",
			request: pz.Request{
				Body: strings.NewReader(url.Values{}.Encode()),
			},
			route: mainHandlerRoute,
			flowParams: confirmationFlowParams{
				activity: "Activity",
				basePath: "/activity",
				fields:   []field{{}},
				mainCallback: func(
					auth *AuthService,
					form url.Values,
				) (types.UserID, error) {
					return "", nil
				},
			},
			wantedStatus: http.StatusAccepted,
			wantedBody:   testsupport.WantedString(Must(ackPage("Activity"))),
		},
		{
			name:    "main-handler-invalid-request-body",
			request: pz.Request{Body: strings.NewReader(";")},
			route:   mainHandlerRoute,
			flowParams: confirmationFlowParams{
				activity: "Activity",
				basePath: "/activity",
				fields:   []field{{}},
				mainCallback: func(
					*AuthService,
					url.Values,
				) (types.UserID, error) {
					return "", nil
				},
			},
			wantedStatus: http.StatusBadRequest,
			wantedBody: wantedFormDocument(&formDocument{
				title:        "Activity",
				action:       "/activity",
				errorMessage: "error parsing form data",
			}),
		},
		{
			name:    "main-handler-callback-error-500",
			request: pz.Request{Body: strings.NewReader("")},
			route:   mainHandlerRoute,
			flowParams: confirmationFlowParams{
				activity: "Activity",
				basePath: "/activity",
				fields:   nil,
				mainCallback: func(
					*AuthService,
					url.Values,
				) (types.UserID, error) {
					return "", fmt.Errorf("TEST ERROR")
				},
			},
			wantedStatus: http.StatusInternalServerError,
			wantedBody: wantedFormDocument(&formDocument{
				title:        "Activity",
				action:       "/activity",
				errorMessage: "internal server error",
			}),
		},
		{
			name:    "main-handler-callback-error-inherited",
			request: pz.Request{Body: strings.NewReader("")},
			route:   mainHandlerRoute,
			flowParams: confirmationFlowParams{
				activity: "Activity",
				basePath: "/activity",
				fields:   nil,
				mainCallback: func(
					*AuthService,
					url.Values,
				) (types.UserID, error) {
					return "", &pz.HTTPError{Status: 700, Message: "TEST ERR"}
				},
			},
			wantedStatus: 700,
			wantedBody: wantedFormDocument(&formDocument{
				title:        "Activity",
				action:       "/activity",
				errorMessage: "TEST ERR",
			}),
		},
		{
			// When the callback returns a `*successSentinelErr`, the `success`
			// callback should be invoked.
			name:    "main-handler-callback-error-sentinel",
			request: pz.Request{Body: strings.NewReader("")},
			route:   mainHandlerRoute,
			flowParams: confirmationFlowParams{
				activity: "Activity",
				basePath: "/activity",
				fields:   nil,
				mainCallback: func(
					*AuthService,
					url.Values,
				) (types.UserID, error) {
					return "", &successSentinelErr{
						message: "TEST ERROR",
						err:     fmt.Errorf("TEST ERROR"),
					}
				},
			},
			wantedStatus: http.StatusAccepted,
			wantedBody:   testsupport.WantedString(Must(ackPage("Activity"))),
		},
		{
			name: "confirmation-form",
			request: pz.Request{
				URL: Must(url.Parse("https://example.org?t=token")),
			},
			route: confirmationFormRoute,
			flowParams: confirmationFlowParams{
				activity: "Activity",
				basePath: "/activity",
			},
			wantedStatus: http.StatusOK,
			wantedBody: wantedFormDocument(&formDocument{
				title:  "Confirm Activity",
				action: "/activity/confirm",
				fields: []input{{
					name:  "password",
					type_: "password",
					value: "",
					label: "Password",
				}, {
					name:  "token",
					type_: "hidden",
					value: "token",
					label: "",
				}},
			}),
		},
		{
			name: "confirmation-handler-simple-create",
			request: pz.Request{
				Body: strings.NewReader(encodeForm(
					kv{"token", testsupport.ResetToken},
					kv{"password", testsupport.GoodPassword},
				)),
			},
			route: confirmationHandlerRoute,
			flowParams: confirmationFlowParams{
				activity: "Activity",
				basePath: "/activity",
				fields:   []field{{}},
				create:   true,
			},
			wantedUsers: []types.Credentials{{
				User:     "user",
				Email:    "user@example.org",
				Password: testsupport.GoodPassword,
			}},
			wantedStatus: http.StatusSeeOther,
			wantedBody:   testsupport.WantedString("303 See Other"),
		},
		{
			name: "confirmation-handler-simple-no-create",
			request: pz.Request{
				Body: strings.NewReader(encodeForm(
					kv{"token", testsupport.ResetToken},
					kv{"password", testsupport.GoodPassword},
				)),
			},
			route: confirmationHandlerRoute,
			flowParams: confirmationFlowParams{
				activity: "Activity",
				basePath: "/activity",
				fields:   []field{{}},
				create:   false,
			},
			users: testsupport.UserStoreFake{
				"user": &types.UserEntry{
					User:  "user",
					Email: "user@example.org",
				},
			},
			wantedStatus: http.StatusSeeOther,
			wantedBody:   testsupport.WantedString("303 See Other"),
			wantedUsers: []types.Credentials{{
				User:     "user",
				Email:    "user@example.org",
				Password: testsupport.GoodPassword,
			}},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			flow, err := newConfirmationFlow(&testCase.flowParams)
			if err != nil {
				t.Fatalf(
					"unexpected error building `confirmationFlow`: %v",
					err,
				)
			}
			if testCase.users == nil {
				testCase.users = testsupport.UserStoreFake{}
			}
			rsp := testCase.route(
				testWebServer(testCase.users, nil),
				flow,
			).Handler(testCase.request)

			if rsp.Status != testCase.wantedStatus {
				t.Fatalf(
					"wanted status `%d`; found `%d`",
					testCase.wantedStatus,
					rsp.Status,
				)
			}

			if err := testCase.wantedBody(rsp.Data); err != nil {
				t.Fatal(err)
			}

			if err := testCase.users.ExpectUsers(
				testCase.wantedUsers,
			); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func testWebServer(
	users types.UserStore,
	notifications types.NotificationService,
) *WebServer {
	return &WebServer{
		AuthService:             *testAuthService(users, notifications),
		BaseURL:                 testsupport.BaseURL,
		RedirectDomain:          testsupport.RedirectDomain,
		DefaultRedirectLocation: testsupport.DefaultRedirectLocation,
	}
}

func testAuthService(
	users types.UserStore,
	notifications types.NotificationService,
) *AuthService {
	return &AuthService{
		Creds:         CredStore{Users: users},
		ResetTokens:   testsupport.ResetTokenFactory,
		Tokens:        testsupport.TokenStoreFake{},
		Notifications: notifications,
		Codes:         testsupport.CodesTokenFactory,
		TokenDetails: TokenDetailsFactory{
			AccessTokens:  testsupport.AccessTokenFactory,
			RefreshTokens: testsupport.RefreshTokenFactory,
			TimeFunc:      testsupport.NowTimeFunc,
		},
		TimeFunc: testsupport.NowTimeFunc,
	}
}

func mainFormRoute(ws *WebServer, flow *confirmationFlow) pz.Route {
	return flow.routeMainForm()
}

func confirmationFormRoute(ws *WebServer, flow *confirmationFlow) pz.Route {
	return flow.routeConfirmationForm()
}

func mainHandlerRoute(ws *WebServer, flow *confirmationFlow) pz.Route {
	return flow.routeMainHandler(ws)
}

func confirmationHandlerRoute(ws *WebServer, flow *confirmationFlow) pz.Route {
	return flow.routeConfirmationHandler(ws)
}
