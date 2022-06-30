package client

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	pz "github.com/weberc2/httpeasy"
	pztest "github.com/weberc2/httpeasy/testsupport"
	"github.com/weberc2/mono/pkg/auth"
	"github.com/weberc2/mono/pkg/auth/testsupport"
	"github.com/weberc2/mono/pkg/auth/types"
	"golang.org/x/crypto/bcrypt"
)

func TestClient_Logout(t *testing.T) {
	jwt.TimeFunc = func() time.Time { return now }
	defer func() { jwt.TimeFunc = time.Now }()
	users := testsupport.UserStoreFake{
		"user": &types.UserEntry{
			User:  "user",
			Email: "user@example.org",
			PasswordHash: func() []byte {
				hash, err := bcrypt.GenerateFromPassword(
					[]byte("password"),
					bcrypt.DefaultCost,
				)
				if err != nil {
					t.Fatalf(
						"unexpected error bcrypt-hashing password: %v",
						err,
					)
				}
				return hash
			}(),
		},
	}
	authService, err := testAuthService(&authServiceOptions{userStore: users})
	if err != nil {
		t.Fatalf("creating test `auth.AuthService`: %v", err)
	}
	tokens, err := authService.Login(&types.Credentials{
		User:     "user",
		Email:    "user@example.org",
		Password: "password",
	})
	if err != nil {
		t.Fatalf("unexpected error logging in: %v", err)
	}

	api := auth.AuthHTTPService{AuthService: authService}
	srv := httptest.NewServer(pz.Register(
		pztest.TestLog(t),
		api.RefreshRoute(),
		api.LogoutRoute(),
	))
	defer srv.Close()

	t.Logf("URL: %s", srv.URL)

	client := testClient(srv)

	// make sure we can refresh
	_, err = client.Refresh(tokens.RefreshToken.Token)
	if err != nil {
		t.Fatalf("unexpected error refreshing token: %v", err)
	}

	// logout; make sure there's no error
	if err := client.Logout(tokens.RefreshToken.Token); err != nil {
		t.Fatalf("logout error: expected `nil`; found `%v`", err)
	}

	// make sure we CANNOT refresh
	_, err = client.Refresh(tokens.RefreshToken.Token)
	if err := auth.ErrUnauthorized.CompareErr(err); err != nil {
		t.Fatal(err)
	}
}

func TestClient_Exchange(t *testing.T) {
	for _, testCase := range []struct {
		name         string
		tokenCreated time.Time
		wantedErr    types.WantedError
		wantedTokens bool
	}{
		{
			name:         "simple",
			tokenCreated: now,
			wantedTokens: true,
		},
		{
			name:         "unauthorized",
			tokenCreated: now.Add(-24 * 30 * time.Hour),
			wantedErr:    auth.ErrUnauthorized,
			wantedTokens: false,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			jwt.TimeFunc = func() time.Time { return now }
			defer func() { jwt.TimeFunc = time.Now }()

			authService, err := testAuthService(nil)
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

			code, err := authService.Codes.Create(
				testCase.tokenCreated,
				"adam",
			)
			if err != nil {
				t.Fatalf("unexpected error creating auth code: %v", err)
			}

			if testCase.wantedErr == nil {
				testCase.wantedErr = types.NilError{}
			}
			tokens, err := client.Exchange(code.Token)
			if err := testCase.wantedErr.CompareErr(err); err != nil {
				t.Fatal(err)
			}

			if testCase.wantedTokens && tokens == nil {
				t.Fatal("wanted tokens, but found none")
			}
			if !testCase.wantedTokens && tokens != nil {
				t.Fatal("didn't want tokens, but found tokens")
			}
		})
	}
}

func testClient(srv *httptest.Server) Client {
	return Client{HTTP: *testHTTPClient(srv), BaseURL: srv.URL}
}

func testAuthService(options *authServiceOptions) (auth.AuthService, error) {
	if options == nil {
		options = defaultAuthServiceOptions()
	} else {
		if options.userStore == nil {
			options.userStore = testsupport.UserStoreFake{}
		}
		if options.authCodeFactory == nil {
			options.authCodeFactory = defaultAuthCodeFactory()
		}
		if options.tokenStore == nil {
			options.tokenStore = testsupport.TokenStoreFake{}
		}
	}
	accessKey, err := p521Key()
	if err != nil {
		return auth.AuthService{}, fmt.Errorf(
			"unexpected error generating access key: %w",
			err,
		)
	}
	refreshKey, err := p521Key()
	if err != nil {
		return auth.AuthService{}, fmt.Errorf(
			"unexpected error generating refresh key: %w",
			err,
		)
	}

	return auth.AuthService{
		Tokens:        options.tokenStore,
		Creds:         auth.CredStore{Users: options.userStore},
		Notifications: &testsupport.NotificationServiceFake{},
		Codes:         *options.authCodeFactory,
		TokenDetails: auth.TokenDetailsFactory{
			AccessTokens: types.TokenFactory{
				Issuer:        "issuer",
				Audience:      "audience",
				TokenValidity: 15 * time.Minute,
				SigningKey:    accessKey,
			},
			RefreshTokens: types.TokenFactory{
				Issuer:        "issuer",
				Audience:      "audience",
				TokenValidity: 15 * time.Minute,
				SigningKey:    refreshKey,
			},
			TimeFunc: func() time.Time { return now },
		},
	}, nil
}

type authServiceOptions struct {
	userStore       testsupport.UserStoreFake
	authCodeFactory *types.TokenFactory
	tokenStore      testsupport.TokenStoreFake
}

func defaultAuthCodeFactory() *types.TokenFactory {
	return &types.TokenFactory{
		Issuer:        "issuer",
		Audience:      "audience",
		TokenValidity: time.Minute,
		SigningKey:    mustP521Key(),
	}
}

func defaultAuthServiceOptions() *authServiceOptions {
	return &authServiceOptions{
		userStore:       testsupport.UserStoreFake{},
		authCodeFactory: defaultAuthCodeFactory(),
		tokenStore:      testsupport.TokenStoreFake{},
	}
}

func p521Key() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
}

func mustP521Key() *ecdsa.PrivateKey {
	key, err := p521Key()
	if err != nil {
		panic(fmt.Sprintf("generating ECDSA key: %v", err))
	}
	return key
}

var now = time.Date(1988, 8, 3, 0, 0, 0, 0, time.UTC)
