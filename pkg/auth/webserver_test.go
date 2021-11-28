package auth

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/weberc2/auth/pkg/testsupport"
	pz "github.com/weberc2/httpeasy"
)

func TestLoginHandler(t *testing.T) {
	now := time.Date(1988, 9, 3, 0, 0, 0, 0, time.UTC)
	codes := TokenFactory{
		Issuer:        "issuer",
		Audience:      "audience",
		TokenValidity: time.Minute,
		ParseKey:      []byte("signing-key"),  // symmetric
		SigningKey:    []byte("signing-key"),  // symmetric
		SigningMethod: jwt.SigningMethodHS512, // symmetric, deterministic
	}

	for _, testCase := range []struct {
		name           string
		username       string
		password       string
		redirect       string
		stateUsers     testsupport.UserStoreFake
		wantedStatus   int
		wantedLocation string
	}{{
		name:     "redirects with auth code",
		username: "adam",
		password: "password",
		redirect: "https://app.example.org/users/adam/settings",
		stateUsers: testsupport.UserStoreFake{
			"adam": {
				User:         "adam",
				Email:        "adam@example.org",
				PasswordHash: hashBcrypt("password"),
			},
		},
		wantedStatus: http.StatusSeeOther,
		wantedLocation: fmt.Sprintf(
			"https://app.example.org/users/adam/settings?%s",
			url.Values{
				"code": []string{mustString(codes.Create(now, "adam"))},
			}.Encode(),
		),
	}} {
		webServer := WebServer{
			AuthService: AuthService{
				Creds: CredStore{
					Users: testCase.stateUsers,
				},
				Notifications: testsupport.NotificationServiceFake{},
				Codes:         codes,
				TimeFunc:      func() time.Time { return now },
			},
			BaseURL:                 "https://auth.example.org",
			RedirectDomain:          "app.example.org",
			DefaultRedirectLocation: "https://app.example.org/default/",
		}

		rsp := webServer.LoginHandler(pz.Request{
			Body: strings.NewReader(
				url.Values{
					"username": []string{testCase.username},
					"password": []string{testCase.password},
				}.Encode(),
			),
			URL: &url.URL{
				RawQuery: fmt.Sprintf("location=%s", testCase.redirect),
			},
		})

		if rsp.Status != testCase.wantedStatus {
			t.Fatalf(
				"Response.Status: wanted `%d`; found `%d`",
				testCase.wantedStatus,
				rsp.Status,
			)
		}

		if loc := rsp.Headers.Get("location"); loc != testCase.wantedLocation {
			t.Fatalf(
				"Response.Headers[\"Location\"]: wanted `%s`; found `%s`",
				testCase.wantedLocation,
				loc,
			)
		}
	}
}

func mustString(s string, err error) string {
	if err != nil {
		panic(err)
	}
	return s
}
