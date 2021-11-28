package client

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/weberc2/auth/pkg/auth"
	pz "github.com/weberc2/httpeasy"
	pztest "github.com/weberc2/httpeasy/testsupport"
)

func TestAuthCodeCallback(t *testing.T) {
	now := time.Date(1988, 8, 3, 0, 0, 0, 0, time.UTC)
	jwt.TimeFunc = func() time.Time { return now }
	defer func() { jwt.TimeFunc = time.Now }()

	authService, err := testAuthService(now)
	if err != nil {
		t.Fatalf("creating test `auth.AuthService`: %v", err)
	}

	authSrv := httptest.NewServer(pz.Register(
		pztest.TestLog(t),
		(&auth.AuthHTTPService{AuthService: authService}).ExchangeRoute(),
	))

	app := WebServerApp{
		Client:          testClient(authSrv),
		BaseURL:         authSrv.URL,
		DefaultRedirect: "default",
		Key:             "cookie-encryption-key",
	}

	appSrv := httptest.NewServer(pz.Register(
		pztest.TestLog(t),
		app.AuthCodeCallbackRoute("/api/auth/code"),
	))

	appClient := appSrv.Client()

	code, err := authService.Codes.Create(now, "adam")
	if err != nil {
		t.Fatalf("unexpected error creating auth code: %v", err)
	}

	rsp, err := appClient.Get(
		fmt.Sprintf(
			"%s/api/auth/code?%s",
			appSrv.URL,
			url.Values{
				"code":     []string{code},
				"redirect": []string{"intended"},
			}.Encode(),
		),
	)
	if err != nil {
		t.Fatalf("unexpected error communicating with app server: %v", err)
	}

	if rsp.StatusCode != http.StatusSeeOther {
		t.Fatalf(
			"Response.StatusCode: wanted `%d`; found `%d`",
			http.StatusSeeOther,
			rsp.StatusCode,
		)
	}

	wanted := "https://app.example.org/intended"
	if rsp.Header.Get("Location") != wanted {
		t.Fatalf(
			"Response.Header[\"Location\"]: wanted `%s`; found `%s`",
			wanted,
			rsp.Header.Get("Location"),
		)
	}

	cookies := rsp.Cookies()
	var accessToken, refreshToken string
	for _, cookie := range cookies {
		if cookie.Name == "Access-Token" {
			accessToken = cookie.Value
		} else if cookie.Name == "Refresh-Token" {
			refreshToken = cookie.Value
		}
	}
	if accessToken == "" {
		t.Fatal("missing `Access-Token` cookie")
	}
	if refreshToken == "" {
		t.Fatal("missing `Refresh-Token` cookie")
	}
}
