package auth

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	pz "github.com/weberc2/httpeasy"
	pztest "github.com/weberc2/httpeasy/testsupport"
	"github.com/weberc2/mono/pkg/auth/testsupport"
	"github.com/weberc2/mono/pkg/auth/types"
)

func TestWebServer_PasswordResetHandlerRoute(t *testing.T) {
	for _, testCase := range []struct {
		name                string
		body                string
		existingUsers       testsupport.UserStoreFake
		wantedStatus        int
		wantedBody          func(pz.Serializer) error
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

			if err := testCase.wantedBody(rsp.Data); err != nil {
				t.Fatalf("Response.Data: %v", err)
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

func wantedHTML(
	callback func(*goquery.Document) error,
) func(pz.Serializer) error {
	return func(s pz.Serializer) error {
		data, err := pztest.ReadAll(s)
		if err != nil {
			return fmt.Errorf("reading serializer: %w", err)
		}

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
