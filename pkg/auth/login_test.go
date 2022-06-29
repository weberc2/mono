package auth

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/dgrijalva/jwt-go"
	pz "github.com/weberc2/httpeasy"
	pztest "github.com/weberc2/httpeasy/testsupport"
	"github.com/weberc2/mono/pkg/auth/testsupport"
	. "github.com/weberc2/mono/pkg/prelude"
)

func TestWebServer_LoginHandler(t *testing.T) {
	for _, testCase := range []struct {
		name           string
		username       string
		password       string
		callback       string
		redirect       string
		stateUsers     testsupport.UserStoreFake
		wantedStatus   int
		wantedBody     func(*goquery.Document) error
		wantedLocation *wantedLocation
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
				PasswordHash: Must(testsupport.HashBcrypt("password")),
			},
		},
		wantedStatus: http.StatusSeeOther,
		wantedLocation: &wantedLocation{
			key:      &testsupport.CodesSigningKey.PublicKey,
			scheme:   "https",
			host:     "app.example.org",
			path:     "/auth/callback",
			callback: "https://app.example.org/auth/callback",
			redirect: "https://app.example.org/users/adam/settings",
		},
	}, {
		name:     "failure redirects to login page",
		username: "adam",
		password: "wrong password",
		callback: "https://app.example.org/auth/callback",
		redirect: "https://app.example.org/users/adam/settings",
		stateUsers: testsupport.UserStoreFake{
			"adam": {
				User:         "adam",
				Email:        "adam@example.org",
				PasswordHash: Must(testsupport.HashBcrypt("password")),
			},
		},
		wantedBody: func(d *goquery.Document) error {
			form := d.Find("form")
			if form.Length() < 1 {
				html, err := d.Html()
				if err != nil {
					return fmt.Errorf(
						"returned document has no `<form>` element (error "+
							"rendering provided HTML: %v)",
						err,
					)
				}
				return fmt.Errorf(
					"returned document has no `<form>` element:\n\n%s",
					html,
				)
			}
			action, exists := form.First().Attr("action")
			if !exists {
				return fmt.Errorf("form has no `action` attribute")
			}

			u, err := url.Parse(action)
			if err != nil {
				return fmt.Errorf("parsing form action: %w", err)
			}
			wantedCallback := "https://app.example.org/auth/callback"
			if cb := u.Query().Get("callback"); cb != wantedCallback {
				return fmt.Errorf(
					"verifying form action: wanted `callback=%s`; found "+
						"`callback=%s`",
					wantedCallback,
					cb,
				)
			}

			return nil
		},
		wantedStatus: http.StatusUnauthorized,
	}} {
		jwt.TimeFunc = testsupport.NowTimeFunc
		defer func() { jwt.TimeFunc = time.Now }()
		webServer := testWebServer(testCase.stateUsers, nil)
		rsp := webServer.LoginHandlerRoute().Handler(pz.Request{
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

		if testCase.wantedBody != nil {
			data, err := pztest.ReadAll(rsp.Data)
			if err != nil {
				t.Fatalf("reading response body: %v", err)
			}
			d, err := goquery.NewDocumentFromReader(bytes.NewReader(data))
			if err != nil {
				t.Fatalf(
					"creating an HTML document from response body: %v",
					err,
				)
			}
			if err := testCase.wantedBody(d); err != nil {
				t.Logf("HTML:\n\n%s", data)
				t.Fatalf("verifying response body: %v", err)
			}
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
	if wanted == nil {
		if found == "" {
			return nil
		}
		return fmt.Errorf(
			"wanted no (or empty) `Location` header; found `%s`",
			found,
		)
	}
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
