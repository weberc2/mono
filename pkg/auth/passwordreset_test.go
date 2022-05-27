package auth

import (
	"bytes"
	"fmt"
	html "html/template"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	pz "github.com/weberc2/httpeasy"
	pztest "github.com/weberc2/httpeasy/testsupport"
	"github.com/weberc2/mono/pkg/auth/testsupport"
	"github.com/weberc2/mono/pkg/auth/types"
	"golang.org/x/crypto/bcrypt"
)

func TestWebServer_PasswordResetConfirmationFormRoute(t *testing.T) {
	logs, err := testFormRoute(
		(*WebServer).PasswordResetConfirmationFormRoute,
		templatePasswordResetConfirmationForm,
		struct {
			FormAction   string
			Token        string
			ErrorMessage string
		}{
			FormAction: pathPasswordResetConfirmation,
		},
	)
	for _, log := range logs {
		t.Log(log)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestWebServer_PasswordResetConfirmationHandlerRoute(t *testing.T) {
	for _, testCase := range []struct {
		name          string
		body          string
		existingUsers testsupport.UserStoreFake
		wantedStatus  int
		wantedUsers   []types.Credentials
	}{
		{
			name: "success",
			body: url.Values{
				"token": []string{mustResetToken(
					t,
					"user",
					"user@example.org",
				)},
				"password": []string{goodPassword},
			}.Encode(),
			existingUsers: testsupport.UserStoreFake{
				"user": &types.UserEntry{
					User:    "user",
					Email:   "user@example.org",
					Created: now,
					PasswordHash: func() []byte {
						hash, err := bcrypt.GenerateFromPassword(
							[]byte("password"),
							bcrypt.DefaultCost,
						)
						if err != nil {
							t.Fatal("failed to hash password: password")
						}
						return hash
					}(),
				},
			},
			wantedStatus: http.StatusSeeOther,
			wantedUsers: []types.Credentials{{
				User:     "user",
				Email:    "user@example.org",
				Password: goodPassword,
			}},
		},
		{
			name: "bad token",
			body: url.Values{
				"token":    []string{"bad token"},
				"password": []string{goodPassword},
			}.Encode(),
			existingUsers: testsupport.UserStoreFake{},
			wantedStatus:  http.StatusBadRequest,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			webServer := WebServer{
				AuthService: AuthService{
					Notifications: &testsupport.NotificationServiceFake{},
					Creds:         CredStore{testCase.existingUsers},
					Tokens:        testsupport.TokenStoreFake{},
					ResetTokens:   resetTokenFactory,
					TimeFunc:      nowTimeFunc,
				},
				BaseURL: "https://auth.example.org",
			}

			rsp := webServer.PasswordResetConfirmationHandlerRoute().Handler(
				pz.Request{Body: strings.NewReader(testCase.body)},
			)

			if rsp.Status != testCase.wantedStatus {
				if data, err := pztest.ReadAll(rsp.Data); err != nil {
					t.Logf(
						"error getting response body for error context: %v",
						err,
					)
				} else {
					t.Logf("Body: %s", data)
				}
				t.Fatalf(
					"Response.Status: wanted `%d`; found `%d`",
					testCase.wantedStatus,
					rsp.Status,
				)
			}

			if err := testCase.existingUsers.ExpectUsers(
				testCase.wantedUsers,
			); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestWebServer_PasswordResetFormRoute(t *testing.T) {
	logs, err := testFormRoute(
		(*WebServer).PasswordResetFormRoute,
		templatePasswordResetForm,
		struct {
			FormAction   string
			ErrorMessage string
		}{
			FormAction: pathPasswordReset,
		},
	)
	for _, log := range logs {
		t.Log(log)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestWebServer_PasswordResetHandlerRoute(t *testing.T) {
	for _, testCase := range []struct {
		name                string
		body                string
		existingUsers       testsupport.UserStoreFake
		wantedStatus        int
		wantedBody          func([]byte) error
		wantedNotifications []*types.Notification
	}{
		{
			name: "success",
			body: url.Values{"username": []string{"user"}}.Encode(),
			existingUsers: testsupport.UserStoreFake{
				"user": &types.UserEntry{
					User:         "user",
					Email:        "user@example.org",
					Created:      now,
					PasswordHash: nil,
				},
			},
			wantedStatus: http.StatusAccepted,
			wantedBody: wantedHTML(func(doc *goquery.Document) error {
				wanted := "Initiated Password Reset"
				if h1 := doc.Find("h1").Text(); wanted != h1 {
					return fmt.Errorf(
						"h1: wanted `%s`; found `%s`",
						wanted,
						h1,
					)
				}
				return nil
			}),
			wantedNotifications: []*types.Notification{{
				Type:  types.NotificationTypeForgotPassword,
				User:  "user",
				Email: "user@example.org",
				Token: mustResetToken(t, "user", "user@example.org"),
			}},
		},
		{
			name:          "user not found",
			body:          url.Values{"username": []string{"user"}}.Encode(),
			existingUsers: testsupport.UserStoreFake{},
			wantedStatus:  http.StatusAccepted,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			notifications := testsupport.NotificationServiceFake{}
			webServer := WebServer{
				AuthService: AuthService{
					Notifications: &notifications,
					Creds:         CredStore{testCase.existingUsers},
					Tokens:        testsupport.TokenStoreFake{},
					ResetTokens:   resetTokenFactory,
					TimeFunc:      nowTimeFunc,
				},
				BaseURL: "https://auth.example.org",
			}

			rsp := webServer.PasswordResetHandlerRoute().Handler(pz.Request{
				Body: strings.NewReader(testCase.body),
			})

			if rsp.Status != testCase.wantedStatus {
				if data, err := pztest.ReadAll(rsp.Data); err != nil {
					t.Logf(
						"error getting response body for error context: %v",
						err,
					)
				} else {
					t.Logf("Body: %s", data)
				}
				t.Fatalf(
					"Response.Status: wanted `%d`; found `%d`",
					testCase.wantedStatus,
					rsp.Status,
				)
			}

			if testCase.wantedBody != nil {
				data, err := pztest.ReadAll(rsp.Data)
				if err != nil {
					t.Fatalf(
						"Response.Data: reading serializer: %v",
						err,
					)
				}

				if err := testCase.wantedBody(data); err != nil {
					t.Fatalf("Response.Data: %v", err)
				}
			}

			if err := compareNotifications(
				resetTokenFactory.SigningKey.PublicKey,
				testCase.wantedNotifications,
				notifications.Notifications,
			); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func testFormRoute(
	route func(*WebServer) pz.Route,
	wantedTemplate *html.Template,
	wantedTemplateParams interface{},
) ([]string, error) {
	notifications := testsupport.NotificationServiceFake{}
	webServer := WebServer{
		AuthService: AuthService{
			Notifications: &notifications,
			Creds:         CredStore{testsupport.UserStoreFake{}},
			Tokens:        testsupport.TokenStoreFake{},
			ResetTokens:   resetTokenFactory,
			TimeFunc:      nowTimeFunc,
		},
		BaseURL: "https://auth.example.org",
	}

	rsp := route(&webServer).Handler(pz.Request{
		Body: strings.NewReader(""),
		URL:  &url.URL{},
	})

	if rsp.Status != http.StatusOK {
		var logs []string
		if data, err := pztest.ReadAll(rsp.Data); err != nil {
			logs = append(logs, fmt.Sprintf(
				"error getting response body for error context: %v",
				err,
			))
		} else {
			logs = append(logs, fmt.Sprintf("Body: %s", data))
		}
		return logs, fmt.Errorf(
			"Response.Status: wanted `202`; found `%d`",
			rsp.Status,
		)
	}

	data, err := pztest.ReadAll(rsp.Data)
	if err != nil {
		return nil, fmt.Errorf("Response.Data: reading serializer: %w", err)
	}

	var sb strings.Builder
	if err := wantedTemplate.Execute(&sb, wantedTemplateParams); err != nil {
		return nil, fmt.Errorf("unexpected error executing template: %w", err)
	}
	if stringData := string(data); stringData != sb.String() {
		return nil, fmt.Errorf(
			"Response.Data: wanted `%s`; found `%s`",
			sb.String(),
			stringData,
		)
	}

	return nil, nil
}

func wantedHTML(
	callback func(*goquery.Document) error,
) func([]byte) error {
	return func(data []byte) error {
		d, err := goquery.NewDocumentFromReader(bytes.NewReader(data))
		if err != nil {
			return fmt.Errorf(
				"building HTML document from serializer: %w",
				err,
			)
		}

		return callback(d)
	}
}
