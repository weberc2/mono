package client

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/weberc2/auth/pkg/auth"
	"github.com/weberc2/auth/pkg/testsupport"
	pz "github.com/weberc2/httpeasy"
	pztest "github.com/weberc2/httpeasy/testsupport"
)

func testClient(srv *httptest.Server) Client {
	httpClient := srv.Client()
	httpClient.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return Client{HTTP: *httpClient, BaseURL: srv.URL}
}

func testAuthService(now time.Time) (auth.AuthService, error) {
	authCodeKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return auth.AuthService{}, fmt.Errorf(
			"unexpected error generating auth code key: %w",
			err,
		)
	}
	accessKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return auth.AuthService{}, fmt.Errorf(
			"unexpected error generating access key: %w",
			err,
		)
	}
	refreshKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return auth.AuthService{}, fmt.Errorf(
			"unexpected error generating refresh key: %w",
			err,
		)
	}

	codes := auth.TokenFactory{
		Issuer:        "issuer",
		Audience:      "audience",
		TokenValidity: time.Minute,
		SigningKey:    authCodeKey,
	}

	return auth.AuthService{
		Creds: auth.CredStore{
			Users: testsupport.UserStoreFake{},
		},
		Notifications: testsupport.NotificationServiceFake{},
		Codes:         codes,
		TokenDetails: auth.TokenDetailsFactory{
			AccessTokens: auth.TokenFactory{
				Issuer:        "issuer",
				Audience:      "audience",
				TokenValidity: 15 * time.Minute,
				SigningKey:    accessKey,
			},
			RefreshTokens: auth.TokenFactory{
				Issuer:        "issuer",
				Audience:      "audience",
				TokenValidity: 15 * time.Minute,
				SigningKey:    refreshKey,
			},
			TimeFunc: func() time.Time { return now },
		},
	}, nil
}

func TestExchangeRoute(t *testing.T) {
	now := time.Date(1988, 8, 3, 0, 0, 0, 0, time.UTC)
	jwt.TimeFunc = func() time.Time { return now }
	defer func() { jwt.TimeFunc = time.Now }()

	authService, err := testAuthService(now)
	if err != nil {
		t.Fatalf("creating test `auth.AuthService`: %v", err)
	}

	api := auth.AuthHTTPService{AuthService: authService}
	srv := httptest.NewServer(pz.Register(
		pztest.TestLog(t),
		api.ExchangeRoute(),
	))
	defer srv.Close()

	t.Logf("URL: %s", srv.URL)

	client := testClient(srv)

	code, err := authService.Codes.Create(now, "adam")
	if err != nil {
		t.Fatalf("unexpected error creating auth code: %v", err)
	}

	tokens, err := client.Exchange(code.Token)
	if err != nil {
		t.Fatalf("unexpected error exchanging auth code: %v", err)
	}

	if tokens.AccessToken.Token == "" {
		t.Fatal("missing access token")
	}

	if tokens.RefreshToken.Token == "" {
		t.Fatal("missing refresh token")
	}
}
