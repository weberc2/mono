package auth

import (
	"crypto/ecdsa"
	"encoding/json"
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
	codes := TokenFactory{
		Issuer:        "issuer",
		Audience:      "audience",
		TokenValidity: time.Minute,
		SigningKey:    codesSigningKey,
	}

	for _, testCase := range []struct {
		name           string
		username       string
		password       string
		callback       string
		redirect       string
		stateUsers     testsupport.UserStoreFake
		wantedStatus   int
		wantedLocation wantedLocation
	}{{
		name:     "redirects with auth code",
		username: "adam",
		password: "password",
		callback: "https://app.example.org/auth/callback",
		redirect: "https://app.example.org/users/adam/settings",
		stateUsers: testsupport.UserStoreFake{
			"adam": {
				User:         "adam",
				Email:        "adam@example.org",
				PasswordHash: hashBcrypt("password"),
			},
		},
		wantedStatus: http.StatusSeeOther,
		wantedLocation: wantedLocation{
			key:      &codesSigningKey.PublicKey,
			scheme:   "https",
			host:     "app.example.org",
			path:     "/auth/callback",
			callback: "https://app.example.org/auth/callback",
			redirect: "https://app.example.org/users/adam/settings",
		},
	}} {
		jwt.TimeFunc = nowTimeFunc
		defer func() { jwt.TimeFunc = time.Now }()
		webServer := WebServer{
			AuthService: AuthService{
				Creds: CredStore{
					Users: testCase.stateUsers,
				},
				Notifications: testsupport.NotificationServiceFake{},
				Codes:         codes,
				TimeFunc:      nowTimeFunc,
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
				RawQuery: url.Values{
					"redirect": []string{testCase.redirect},
					"callback": []string{testCase.callback},
				}.Encode(),
			},
		})

		if rsp.Status != testCase.wantedStatus {
			data, err := json.Marshal(rsp.Logging)
			if err != nil {
				t.Logf("marshaling response logs: %v", err)
			}
			t.Logf("response logs: %s", data)
			t.Fatalf(
				"Response.Status: wanted `%d`; found `%d`",
				testCase.wantedStatus,
				rsp.Status,
			)
		}

		if err := testCase.wantedLocation.compare(
			rsp.Headers.Get("Location"),
		); err != nil {
			t.Fatalf("Response.Headers[\"Location\"]: %v", err)
		}
	}
}

type wantedLocation struct {
	key      *ecdsa.PublicKey
	scheme   string
	host     string
	path     string
	callback string
	redirect string
}

func (wanted *wantedLocation) compare(found string) error {
	url, err := url.Parse(found)
	if err != nil {
		return fmt.Errorf("parsing `found`: %w", err)
	}

	if wanted.scheme != url.Scheme {
		return fmt.Errorf(
			"URL.Scheme: wanted `%s`; found `%s`",
			wanted.scheme,
			url.Scheme,
		)
	}

	if wanted.host != url.Host {
		return fmt.Errorf(
			"URL.Host: wanted `%s`; found `%s`",
			wanted.host,
			url.Host,
		)
	}

	if wanted.path != url.Path {
		return fmt.Errorf(
			"URL.Path: wanted `%s`; found `%s`",
			wanted.path,
			url.Path,
		)
	}

	query := url.Query()
	if len(query) != 3 {
		return fmt.Errorf("URL.Query: wanted `3` keys; found `%d`", len(query))
	}

	if wanted.callback != query.Get("callback") {
		return fmt.Errorf(
			"URL.Query[\"callback\"]: wanted `%s`; found `%s`",
			wanted.callback,
			query.Get("callback"),
		)
	}

	if wanted.redirect != query.Get("redirect") {
		return fmt.Errorf(
			"URL.Query[\"redirect\"]: wanted `%s`; found `%s`",
			wanted.redirect,
			query.Get("redirect"),
		)
	}

	if _, err := jwt.Parse(
		query.Get("code"),
		func(*jwt.Token) (interface{}, error) {
			return wanted.key, nil
		},
	); err != nil {
		return fmt.Errorf("URL.Query[\"code\"]: parsing JWT: %w", err)
	}

	return nil
}
