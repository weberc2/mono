package auth

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
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
	pztest "github.com/weberc2/httpeasy/testsupport"
)

func TestLoginHandler(t *testing.T) {
	now := time.Date(1988, 9, 3, 0, 0, 0, 0, time.UTC)
	codes := TokenFactory{
		Issuer:           "issuer",
		Audience: "audience",
		TokenValidity:    time.Minute,
		ParseKey:         []byte("signing-key"),  // symmetric
		SigningKey:       []byte("signing-key"),  // symmetric
		SigningMethod:    jwt.SigningMethodHS512, // symmetric, deterministic
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

func TestExchange(t *testing.T) {
	now := time.Date(1988, 9, 3, 0, 0, 0, 0, time.UTC)
	jwt.TimeFunc = func() time.Time { return now }
	defer func() { jwt.TimeFunc = time.Now }()
	codes := TokenFactory{
		Issuer:           "issuer",
		Audience: "audience",
		TokenValidity:    time.Minute,
		ParseKey:         []byte("signing-key"),  // symmetric
		SigningKey:       []byte("signing-key"),  // symmetric
		SigningMethod:    jwt.SigningMethodHS512, // symmetric, deterministic
	}

	for _, testCase := range []struct {
		name         string
		username     string
		code         string
		wantedStatus int
		wantedBody   pztest.WantedData
	}{{
		name:         "good code",
		username:     "adam",
		code:         mustString(codes.Create(now, "adam")),
		wantedStatus: http.StatusOK,
		wantedBody:   AnyTokens{}, // good enough
	}} {
		t.Run(testCase.name, func(t *testing.T) {
			accessKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
			if err != nil {
				t.Fatalf("unexpected error generating access key: %v", err)
			}
			refreshKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
			if err != nil {
				t.Fatalf("unexpected error generating refresh key: %v", err)
			}

			webServer := WebServer{
				AuthService: AuthService{
					Notifications: testsupport.NotificationServiceFake{},
					TokenDetails: TokenDetailsFactory{
						AccessTokens: TokenFactory{
							Issuer:           "issuer",
							Audience: "audience",
							TokenValidity:    15 * time.Minute,
							ParseKey:         &accessKey.PublicKey,
							SigningKey:       accessKey,
							SigningMethod:    jwt.SigningMethodES512,
						},
						RefreshTokens: TokenFactory{
							Issuer:           "issuer",
							Audience: "audience",
							TokenValidity:    7 * 24 * time.Hour,
							ParseKey:         &refreshKey.PublicKey,
							SigningKey:       refreshKey,
							SigningMethod:    jwt.SigningMethodES512,
						},
						TimeFunc: func() time.Time { return now },
					},
					Codes:    codes,
					TimeFunc: func() time.Time { return now },
				},
				BaseURL:                 "https://auth.example.org",
				RedirectDomain:          "app.example.org",
				DefaultRedirectLocation: "https://app.example.org/default/",
			}

			data, err := json.Marshal(struct {
				Code string `json:"code"`
			}{
				testCase.code,
			})
			if err != nil {
				t.Fatalf("unexpected error marshaling auth code: %v", err)
			}

			rsp := webServer.Exchange(
				pz.Request{Body: bytes.NewReader(data)},
			)

			if rsp.Status != testCase.wantedStatus {
				data, err := readAll(rsp.Data)
				if err != nil {
					t.Logf("reading response body: %v", err)
				}
				t.Logf("response body: %s", data)

				data, err = json.Marshal(rsp.Logging)
				if err != nil {
					t.Logf("marshaling response logging: %v", err)
				}
				t.Logf("response logging: %s", data)
				t.Fatalf(
					"Response.Status: wanted `%d`; found `%d`",
					testCase.wantedStatus,
					rsp.Status,
				)
			}

			if err := pztest.CompareSerializer(
				testCase.wantedBody,
				rsp.Data,
			); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func mustString(s string, err error) string {
	if err != nil {
		panic(err)
	}
	return s
}

type AnyTokens struct{}

func (AnyTokens) CompareData(data []byte) error {
	var details TokenDetails

	if err := json.Unmarshal(data, &details); err != nil {
		return fmt.Errorf("unmarshaling `TokenDetails`: %w", err)
	}

	if details.AccessToken == "" {
		return fmt.Errorf("missing access token")
	}

	if details.RefreshToken == "" {
		return fmt.Errorf("missing refresh token")
	}

	return nil
}
