package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/weberc2/mono/pkg/auth/testsupport"
	"github.com/weberc2/mono/pkg/auth/types"
	. "github.com/weberc2/mono/pkg/prelude"
)

type userStoreMock struct {
	get    func(types.UserID) (*types.UserEntry, error)
	insert func(*types.UserEntry) error
}

func (usm *userStoreMock) Get(u types.UserID) (*types.UserEntry, error) {
	if usm.get == nil {
		panic("userStoreMock: missing `get` hook")
	}
	return usm.get(u)
}

func (usm *userStoreMock) Update(entry *types.UserEntry) error {
	panic("`userStoreMock.Update()` method not defined")
}

func (usm *userStoreMock) Insert(entry *types.UserEntry) error {
	if usm.insert == nil {
		panic("userStoreMock: missing `create` hook")
	}
	return usm.insert(entry)
}

func TestAuthService_UpdatePassword(t *testing.T) {
	for _, testCase := range []struct {
		name          string
		subject       types.UserID
		token         string
		email         string
		password      string
		create        bool
		existingUsers testsupport.UserStoreFake
		wanted        *types.Credentials
		wantedErr     types.WantedError
	}{
		{
			name:     "simple",
			subject:  testsupport.User,
			token:    testsupport.ResetToken,
			email:    testsupport.Email,
			password: testsupport.GoodPassword,
			create:   true,
			wanted: &types.Credentials{
				User:     testsupport.User,
				Email:    testsupport.Email,
				Password: testsupport.GoodPassword,
			},
		},
		{
			name:     "token parse err",
			subject:  testsupport.User,
			token:    "",
			email:    testsupport.Email,
			password: testsupport.GoodPassword,
			create:   true,
			wanted:   nil,
			wantedErr: types.InvalidRefreshTokenErr(
				jwt.NewValidationError(
					"token contains an invalid number of segments",
					jwt.ValidationErrorMalformed,
				),
			),
		},
		{
			name:      "password validation err",
			subject:   testsupport.User,
			token:     testsupport.ResetToken,
			email:     testsupport.Email,
			password:  "", // invalid
			create:    true,
			wanted:    nil,
			wantedErr: ErrPasswordTooSimple,
		},
		{
			name:     "update",
			subject:  testsupport.User,
			token:    testsupport.ResetToken,
			email:    testsupport.Email,
			password: testsupport.GoodPassword,
			create:   false,
			existingUsers: testsupport.UserStoreFake{
				testsupport.User: &types.UserEntry{
					User:  testsupport.User,
					Email: testsupport.Email,
				},
			},
			wanted: &types.Credentials{
				User:     testsupport.User,
				Email:    testsupport.Email,
				Password: testsupport.GoodPassword,
			},
			wantedErr: nil,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			jwt.TimeFunc = testsupport.NowTimeFunc
			defer func() { jwt.TimeFunc = time.Now }()
			if testCase.existingUsers == nil {
				testCase.existingUsers = testsupport.UserStoreFake{}
			}

			if testCase.wantedErr == nil {
				testCase.wantedErr = types.NilError{}
			}
			authService := testAuthService(testCase.existingUsers, nil)
			_, err := authService.UpdatePassword(&UpdatePassword{
				true,
				testCase.token,
				testCase.password,
			})
			if err := testCase.wantedErr.CompareErr(err); err != nil {
				t.Fatal(err)
			}

			entry, err := testCase.existingUsers.Get(testCase.subject)
			if testCase.wanted == nil && errors.Is(
				err,
				types.ErrUserNotFound,
			) {
				return
			}
			if err != nil {
				t.Fatalf(
					"unexpected error getting user from user store: %v",
					err,
				)
			}

			if err := testCase.wanted.CompareUserEntry(entry); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	accessKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}
	refreshKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}
	const password = "pass"
	hashed := Must(testsupport.HashBcrypt(password))
	if err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}

	jwt.TimeFunc = func() time.Time {
		return testsupport.Now.Add(1 * time.Second)
	}
	tokenStore := testsupport.TokenStoreFake{}
	authService := AuthService{
		Creds: CredStore{&userStoreMock{
			get: func(u types.UserID) (*types.UserEntry, error) {
				if u != "user" {
					return nil, types.ErrUserNotFound
				}
				return &types.UserEntry{
					User:         "user",
					PasswordHash: hashed,
				}, nil
			},
		}},
		Tokens: tokenStore,
		TokenDetails: TokenDetailsFactory{
			AccessTokens: types.TokenFactory{
				Issuer:        "issuer",
				Audience:      "*.example.org",
				TokenValidity: 15 * time.Minute,
				SigningKey:    accessKey,
			},
			RefreshTokens: types.TokenFactory{
				Issuer:        "issuer",
				Audience:      "*.example.org",
				TokenValidity: 7 * 24 * time.Hour,
				SigningKey:    refreshKey,
			},
			TimeFunc: testsupport.NowTimeFunc,
		},
	}

	tokens, err := authService.Login(&types.Credentials{
		User:     "user",
		Password: password,
	})

	if err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}

	claims, err := parseToken(tokens.AccessToken.Token, accessKey)
	if err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}

	if wanted := (jwt.StandardClaims{
		ExpiresAt: testsupport.Now.Add(15 * time.Minute).Unix(),
		IssuedAt:  testsupport.Now.Unix(),
		NotBefore: testsupport.Now.Unix(),
		Issuer:    "issuer",
		Subject:   "user",
		Audience:  "*.example.org",
	}); wanted != *claims {
		t.Fatalf("Wanted:\n%# v\n\nFound:\n%# v", wanted, claims)
	}

	claims, err = parseToken(tokens.RefreshToken.Token, refreshKey)
	if err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}

	if wanted := (jwt.StandardClaims{
		ExpiresAt: testsupport.Now.Add(7 * 24 * time.Hour).Unix(),
		IssuedAt:  testsupport.Now.Unix(),
		NotBefore: testsupport.Now.Unix(),
		Issuer:    "issuer",
		Subject:   "user",
		Audience:  "*.example.org",
	}); wanted != *claims {
		t.Fatalf("Wanted:\n%# v\n\nFound:\n%# v", wanted, claims)
	}

	// make sure the token was persisted
	entries, _ := tokenStore.List()
	if err := types.CompareTokens(
		[]types.Token{{
			Token:   tokens.RefreshToken.Token,
			Expires: tokens.RefreshToken.Expires,
		}},
		entries,
	); err != nil {
		t.Fatal(err)
	}

}

func parseToken(
	token string,
	key *ecdsa.PrivateKey,
) (*jwt.StandardClaims, error) {
	tok, err := jwt.ParseWithClaims(
		token,
		&jwt.StandardClaims{},
		func(*jwt.Token) (interface{}, error) { return &key.PublicKey, nil },
	)
	if err != nil {
		return nil, fmt.Errorf("parsing token: %w", err)
	}

	if tok.Method != jwt.SigningMethodES512 {
		return nil, fmt.Errorf(
			"wanted method '%v'; found '%v'",
			jwt.SigningMethodES512,
			tok.Method,
		)
	}

	if c, ok := tok.Claims.(*jwt.StandardClaims); ok {
		return c, nil
	}

	return nil, fmt.Errorf("invalid claims type: %T", tok.Claims)
}

type notificationServiceMock struct {
	notify func(*types.Notification) error
}

func (nsm *notificationServiceMock) Notify(rt *types.Notification) error {
	if nsm.notify == nil {
		panic("notificationServiceMock: missing `notify` hook")
	}
	return nsm.notify(rt)
}

func TestAuthService_Register(t *testing.T) {
	var (
		notifyCalledWithToken *types.Notification
	)
	resetKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}
	authService := AuthService{
		Creds: CredStore{&userStoreMock{
			get: func(u types.UserID) (*types.UserEntry, error) {
				return nil, types.ErrUserNotFound
			},
		}},
		ResetTokens: types.ResetTokenFactory{
			Issuer:        "issuer",
			Audience:      "audience",
			TokenValidity: 5 * time.Minute,
			SigningKey:    resetKey,
		},
		Notifications: &notificationServiceMock{
			notify: func(n *types.Notification) error {
				notifyCalledWithToken = n
				return nil
			},
		},
		TimeFunc: testsupport.NowTimeFunc,
	}

	if err := authService.Register("user", "user@example.org"); err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}

	wantedToken := &types.Notification{
		Type:  types.NotificationTypeRegister,
		User:  "user",
		Email: "user@example.org",
		Token: notifyCalledWithToken.Token,
	}

	if err := wantedToken.Compare(notifyCalledWithToken); err != nil {
		t.Fatalf("NotificationService.Notify(*ResetToken): %v", err)
	}
}

func TestAuthService_Register_UserNameExists(t *testing.T) {
	authService := AuthService{
		Creds: CredStore{&userStoreMock{
			get: func(u types.UserID) (*types.UserEntry, error) {
				return &types.UserEntry{
					User:         u,
					Email:        "user@example.org",
					PasswordHash: Must(testsupport.HashBcrypt("password")),
				}, nil
			},
		}},
	}

	if err := authService.Register("user", "user@example.org"); err != nil {
		if !errors.Is(err, ErrUserExists) {
			t.Fatalf(
				"Wanted error '%s'; found '%s'",
				ErrUserExists.Error(),
				err.Error(),
			)
		}
		return
	}
	t.Fatal("Wanted `ErrUserExists`; found `<nil>`")
}

func TestAuthService_Register_InvalidEmailAddress(t *testing.T) {
	authService := AuthService{}
	for _, email := range []string{"", "nodomain@", "noatsign"} {
		if err := authService.Register("user", email); err != nil {
			if errors.Is(err, ErrInvalidEmail) {
				continue
			}
			t.Fatalf("Unexpected err: %v", err)
		}
		t.Fatal("Wanted `ErrInvalidEmail`; found `<nil>`")
	}
}

func TestAuthService_ForgotPassword(t *testing.T) {
	var (
		getCalledWithUser     types.UserID
		notifyCalledWithToken *types.Notification
		validity              = 5 * time.Minute
	)
	resetKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}
	authService := AuthService{
		Creds: CredStore{
			Users: &userStoreMock{
				get: func(u types.UserID) (*types.UserEntry, error) {
					getCalledWithUser = u
					return &types.UserEntry{
						User:  u,
						Email: "user@example.org",
					}, nil
				},
			},
		},
		Notifications: &notificationServiceMock{
			notify: func(n *types.Notification) error {
				notifyCalledWithToken = n
				return nil
			},
		},
		ResetTokens: types.ResetTokenFactory{
			Issuer:        "issuer",
			Audience:      "audience",
			TokenValidity: validity,
			SigningKey:    resetKey,
		},
		TimeFunc: testsupport.NowTimeFunc,
	}

	if err := authService.ForgotPassword("user"); err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}

	if getCalledWithUser != "user" {
		t.Fatalf(
			"UserStore.Get(): wanted 'user'; found '%s'",
			getCalledWithUser,
		)
	}

	wantedToken := &types.Notification{
		Type:  types.NotificationTypeForgotPassword,
		User:  "user",
		Email: "user@example.org",
		Token: notifyCalledWithToken.Token, // force a match on this field
	}

	if err := wantedToken.Compare(notifyCalledWithToken); err != nil {
		t.Fatalf("NotificationService.Notify(*ResetToken): %v", err)
	}
}

func TestAuthService_Logout(t *testing.T) {
	for _, testCase := range []struct {
		name         string
		state        testsupport.TokenStoreFake
		refreshToken string
		wantedErr    types.WantedError
		wantedState  []types.Token
	}{
		{
			name:         "simple",
			state:        testsupport.TokenStoreFake{"token": testsupport.Now},
			refreshToken: "token",
		},
		{
			name:         "not found",
			state:        testsupport.TokenStoreFake{},
			refreshToken: "token",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.wantedErr == nil {
				testCase.wantedErr = types.NilError{}
			}
			if err := testCase.wantedErr.CompareErr(
				(*AuthService).Logout(
					&AuthService{Tokens: testCase.state},
					testCase.refreshToken,
				),
			); err != nil {
				t.Fatal(err)
			}

			found, _ := testCase.state.List()
			if err := types.CompareTokens(
				testCase.wantedState,
				found,
			); err != nil {
				t.Fatal(err)
			}
		})
	}
}
