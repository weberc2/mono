package client

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/weberc2/auth/pkg/auth"
	"github.com/weberc2/auth/pkg/testsupport"
	pz "github.com/weberc2/httpeasy"
)

func testAuthService(now time.Time) (auth.AuthService, error) {
	authCodeKey := []byte("auth-code-key")
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
		ParseKey:      authCodeKey,
		SigningKey:    authCodeKey,
		SigningMethod: jwt.SigningMethodHS512,
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
				ParseKey:      &accessKey.PublicKey,
				SigningKey:    accessKey,
				SigningMethod: jwt.SigningMethodES512,
			},
			RefreshTokens: auth.TokenFactory{
				Issuer:        "issuer",
				Audience:      "audience",
				TokenValidity: 15 * time.Minute,
				ParseKey:      &refreshKey.PublicKey,
				SigningKey:    refreshKey,
				SigningMethod: jwt.SigningMethodES512,
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
		pz.JSONLog(os.Stderr),
		api.ExchangeRoute(),
	))
	defer srv.Close()

	t.Logf("URL: %s", srv.URL)

	client := Client{HTTP: *srv.Client(), BaseURL: srv.URL}

	code, err := authService.Codes.Create(now, "adam")
	if err != nil {
		t.Fatalf("unexpected error creating auth code: %v", err)
	}

	tokens, err := client.Exchange(code)
	if err != nil {
		t.Fatalf("unexpected error exchanging auth code: %v", err)
	}

	if tokens.AccessToken == "" {
		t.Fatal("missing access token")
	}

	if tokens.RefreshToken == "" {
		t.Fatal("missing refresh token")
	}
}
