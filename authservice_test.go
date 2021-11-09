package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

type userStoreMock struct {
	get    func(UserID) (*UserEntry, error)
	update func(*UserEntry) error
	create func(*UserEntry) error
}

func (usm *userStoreMock) Get(u UserID) (*UserEntry, error) {
	if usm.get == nil {
		panic("userStoreMock: missing `get` hook")
	}
	return usm.get(u)
}

func (usm *userStoreMock) Update(entry *UserEntry) error {
	if usm.update == nil {
		panic("userStoreMock: missing `update` hook")
	}
	return usm.update(entry)
}

func (usm *userStoreMock) Create(entry *UserEntry) error {
	if usm.create == nil {
		panic("userStoreMock: missing `create` hook")
	}
	return usm.create(entry)
}

func TestLogin(t *testing.T) {
	accessKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}
	refreshKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}
	const password = "pass"
	hashed := hashBcrypt(password)
	if err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}

	now := time.Date(2022, 01, 01, 0, 0, 0, 0, time.UTC)
	jwt.TimeFunc = func() time.Time { return now.Add(1 * time.Second) }
	authService := AuthService{
		Creds: CredStore{&userStoreMock{
			get: func(u UserID) (*UserEntry, error) {
				if u != "user" {
					return nil, ErrUserNotFound
				}
				return &UserEntry{User: "user", PasswordHash: hashed}, nil
			},
		}},
		TokenDetails: TokenDetailsFactory{
			AccessTokens: TokenFactory{
				Issuer:           "issuer",
				WildcardAudience: "*.example.org",
				TokenValidity:    15 * time.Minute,
				SigningKey:       accessKey,
				SigningMethod:    jwt.SigningMethodES512,
			},
			RefreshTokens: TokenFactory{
				Issuer:           "issuer",
				WildcardAudience: "*.example.org",
				TokenValidity:    7 * 24 * time.Hour,
				SigningKey:       refreshKey,
				SigningMethod:    jwt.SigningMethodES512,
			},
			TimeFunc: func() time.Time { return now },
		},
	}

	tokens, err := authService.Login(&Credentials{
		User:     "user",
		Password: password,
	})

	if err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}

	claims, err := parseToken(tokens.AccessToken, accessKey)
	if err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}

	if wanted := (jwt.StandardClaims{
		ExpiresAt: now.Add(15 * time.Minute).Unix(),
		IssuedAt:  now.Unix(),
		NotBefore: now.Unix(),
		Issuer:    "issuer",
		Subject:   "user",
		Audience:  "*.example.org",
	}); wanted != *claims {
		t.Fatalf("Wanted:\n%# v\n\nFound:\n%# v", wanted, claims)
	}

	claims, err = parseToken(tokens.RefreshToken, refreshKey)
	if err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}

	if wanted := (jwt.StandardClaims{
		ExpiresAt: now.Add(7 * 24 * time.Hour).Unix(),
		IssuedAt:  now.Unix(),
		NotBefore: now.Unix(),
		Issuer:    "issuer",
		Subject:   "user",
		Audience:  "*.example.org",
	}); wanted != *claims {
		t.Fatalf("Wanted:\n%# v\n\nFound:\n%# v", wanted, claims)
	}
}

func parseToken(token string, key *ecdsa.PrivateKey) (*jwt.StandardClaims, error) {
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
	notify func(*Notification) error
}

func (nsm *notificationServiceMock) Notify(rt *Notification) error {
	if nsm.notify == nil {
		panic("notificationServiceMock: missing `notify` hook")
	}
	return nsm.notify(rt)
}

func TestRegister(t *testing.T) {
	var (
		notifyCalledWithToken *Notification
		now                   = time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	)
	resetKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}
	authService := AuthService{
		Creds: CredStore{
			Users: &userStoreMock{
				get: func(u UserID) (*UserEntry, error) {
					return nil, ErrUserNotFound
				},
			},
		},
		ResetTokens: ResetTokenFactory{
			Issuer:           "issuer",
			WildcardAudience: "audience",
			TokenValidity:    5 * time.Minute,
			SigningKey:       resetKey,
			SigningMethod:    jwt.SigningMethodES512,
		},
		Notifications: &notificationServiceMock{
			notify: func(n *Notification) error {
				notifyCalledWithToken = n
				return nil
			},
		},
		Hostname: "auth.example.org",
		TimeFunc: func() time.Time { return now },
	}

	if err := authService.Register("user", "user@example.org"); err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}

	wantedToken := &Notification{
		Type:  NotificationTypeRegister,
		User:  "user",
		Email: "user@example.org",
		Token: notifyCalledWithToken.Token,
	}

	if err := wantedToken.compare(notifyCalledWithToken); err != nil {
		t.Fatalf("NotificationService.Notify(*ResetToken): %v", err)
	}
}

func TestRegister_UserNameExists(t *testing.T) {
	authService := AuthService{
		Creds: CredStore{
			Users: &userStoreMock{
				get: func(u UserID) (*UserEntry, error) {
					return &UserEntry{
						User:         u,
						Email:        "user@example.org",
						PasswordHash: hashBcrypt("password"),
					}, nil
				},
			},
		},
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

func TestRegister_InvalidEmailAddress(t *testing.T) {
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

func TestForgotPassword(t *testing.T) {
	var (
		getCalledWithUser     UserID
		notifyCalledWithToken *Notification
		now                   = time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
		validity              = 5 * time.Minute
	)
	resetKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}
	authService := AuthService{
		Creds: CredStore{
			Users: &userStoreMock{
				get: func(u UserID) (*UserEntry, error) {
					getCalledWithUser = u
					return &UserEntry{User: u, Email: "user@example.org"}, nil
				},
			},
		},
		Notifications: &notificationServiceMock{
			notify: func(n *Notification) error {
				notifyCalledWithToken = n
				return nil
			},
		},
		ResetTokens: ResetTokenFactory{
			Issuer:           "issuer",
			WildcardAudience: "audience",
			TokenValidity:    validity,
			SigningKey:       resetKey,
			SigningMethod:    jwt.SigningMethodES512,
		},
		TimeFunc: func() time.Time { return now },
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

	wantedToken := &Notification{
		Type:  NotificationTypeForgotPassword,
		User:  "user",
		Email: "user@example.org",
		Token: notifyCalledWithToken.Token, // force a match on this field
	}

	if err := wantedToken.compare(notifyCalledWithToken); err != nil {
		t.Fatalf("NotificationService.Notify(*ResetToken): %v", err)
	}
}

func TestUpdatePassword(t *testing.T) {
	var (
		now      = time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
		password = "osakldflhkjewadfkjsfduIHUHKJGFU"
		entry    *UserEntry
	)
	resetKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}
	authService := AuthService{
		Creds: CredStore{&userStoreMock{
			get: func(UserID) (*UserEntry, error) {
				return &UserEntry{
					User:         "user",
					Email:        "user@example.org",
					PasswordHash: hashBcrypt(password),
				}, nil
			},
			update: func(e *UserEntry) error { entry = e; return nil },
		}},
		Notifications: &notificationServiceMock{
			notify: func(n *Notification) error { return nil },
		},
		ResetTokens: ResetTokenFactory{
			Issuer:           "issuer",
			WildcardAudience: "audience",
			TokenValidity:    5 * time.Minute,
			SigningKey:       resetKey,
			SigningMethod:    jwt.SigningMethodES512,
		},
		TimeFunc: func() time.Time { return now },
	}

	tok, err := authService.ResetTokens.Create(now, "user", "user@example.org")
	if err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}

	if err := authService.UpdatePassword(&UpdatePassword{
		User:     "user",
		Password: password,
		Token:    tok,
	}); err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}

	wantedCredentials := &Credentials{
		User:     "user",
		Email:    "user@example.org",
		Password: password,
	}

	if err := wantedCredentials.compare(entry); err != nil {
		t.Fatalf("CredStore.Update(*Credentials): %v", err)
	}
}

func (wanted *Credentials) compare(found *UserEntry) error {
	if wanted == nil && found == nil {
		return nil
	}

	if wanted != nil && found == nil {
		return fmt.Errorf("unexpected `nil`")
	}

	if wanted == nil && found != nil {
		return fmt.Errorf("wanted `nil`; found not `nil`")
	}

	if wanted.User != found.User {
		return fmt.Errorf(
			"UserEntry.User: wanted `%s`; found `%s`",
			wanted.User,
			found.User,
		)
	}

	if wanted.Email != found.Email {
		return fmt.Errorf(
			"UserEntry.Email: wanted `%s`; found `%s`",
			wanted.Email,
			found.Email,
		)
	}

	if bcrypt.CompareHashAndPassword(
		found.PasswordHash,
		[]byte(wanted.Password),
	) != nil {
		return fmt.Errorf(
			"UserEntry.PasswordHash: hash doesn't match password `%s`",
			wanted.Password,
		)
	}

	return nil
}

func (wanted *Notification) compare(found *Notification) error {
	if wanted == nil && found == nil {
		return nil
	}

	if wanted != nil && found == nil {
		return fmt.Errorf("unexpected `nil`")
	}

	if wanted == nil && found != nil {
		return fmt.Errorf("wanted `nil`; found not `nil`")
	}

	if wanted.Type != found.Type {
		return fmt.Errorf(
			"ResetToken.Type: `%s`; found `%s`",
			wanted.Type,
			found.Type,
		)
	}

	if wanted.User != found.User {
		return fmt.Errorf(
			"ResetToken.User: wanted `%s`; found `%s`",
			wanted.User,
			found.User,
		)
	}

	if wanted.Email != found.Email {
		return fmt.Errorf(
			"ResetToken.Email: wanted `%s`; found `%s`",
			wanted.Email,
			found.Email,
		)
	}

	if wanted.Token != found.Token {
		return fmt.Errorf(
			"ResetToken.Token: wanted `%s`; found `%s`",
			wanted.Token,
			found.Token,
		)
	}

	return nil
}

func hashBcrypt(password string) []byte {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(fmt.Sprintf("bcrypt-hashing password '%s': %v", password, err))
	}
	return hash
}

// SCENARIOS:
//
// * Multiple registrations with the same email address but different usernames
//   (neither registration is completed)
//   - Option 1: the first to confirm wins (second confirmation gets an error)
//   - Option 2: "registration already exists with the given email address"
// * Multiple registrations for the same email address and username (neither
//   registration is completed)
//   - Invalidate the first token and send a new one
// * Multiple registrations with the same username but different email address
//   - Option 1: update the registration with a new email address and send a
//     new token. This would facilitate the "typo in the email address for the
//     original registration" use case.
// * User attempts to register with an email address (different username) that
//   is already taken
//   - Error: "email address exists"
// * User attempts to register with a username (different email address) that
//   is already taken
//   - Error: "username exists"
