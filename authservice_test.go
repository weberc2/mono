package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

type credStoreMock struct {
	validate func(*Credentials) error
	update   func(*Credentials) error
	create   func(*Credentials) error
}

func (csf *credStoreMock) Validate(cs *Credentials) error {
	if csf.validate == nil {
		panic("credStoreMock: missing `validate` hook")
	}
	return csf.validate(cs)
}

func (csf *credStoreMock) Update(cs *Credentials) error {
	if csf.update == nil {
		panic("credStoreMock: missing `update` hook")
	}
	return csf.update(cs)
}

func (csf *credStoreMock) Create(cs *Credentials) error {
	if csf.create == nil {
		panic("credStoreMock: missing `create` hook")
	}
	return csf.create(cs)
}

func TestLogin(t *testing.T) {
	now := time.Date(2022, 01, 01, 0, 0, 0, 0, time.UTC)
	jwt.TimeFunc = func() time.Time { return now.Add(1 * time.Second) }
	authService := AuthService{
		Creds: &credStoreMock{
			validate: func(cs *Credentials) error {
				if cs.Username == "user" && cs.Password == "pass" {
					return nil
				}
				return ErrCredentials
			},
		},
		TokenDetails: TokenDetailsFactory{
			AccessTokens: TokenFactory{
				Issuer:           "issuer",
				WildcardAudience: "*.example.com",
				TokenValidity:    15 * time.Minute,
				SigningKey:       []byte("access-signing-key"),
				SigningMethod:    jwt.SigningMethodHS512,
			},
			RefreshTokens: TokenFactory{
				Issuer:           "issuer",
				WildcardAudience: "*.example.com",
				TokenValidity:    7 * 24 * time.Hour,
				SigningKey:       []byte("refresh-signing-key"),
				SigningMethod:    jwt.SigningMethodHS512,
			},
			TimeFunc: func() time.Time { return now },
		},
	}

	tokens, err := authService.Login(&Credentials{
		Username: "user",
		Password: "pass",
	})

	if err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}

	claims, err := parseToken(tokens.AccessToken, "access-signing-key")
	if err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}

	if wanted := (jwt.StandardClaims{
		ExpiresAt: now.Add(15 * time.Minute).Unix(),
		IssuedAt:  now.Unix(),
		NotBefore: now.Unix(),
		Issuer:    "issuer",
		Subject:   "user",
		Audience:  "*.example.com",
	}); wanted != *claims {
		t.Fatalf("Wanted:\n%# v\n\nFound:\n%# v", wanted, claims)
	}

	claims, err = parseToken(tokens.RefreshToken, "refresh-signing-key")
	if err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}

	if wanted := (jwt.StandardClaims{
		ExpiresAt: now.Add(7 * 24 * time.Hour).Unix(),
		IssuedAt:  now.Unix(),
		NotBefore: now.Unix(),
		Issuer:    "issuer",
		Subject:   "user",
		Audience:  "*.example.com",
	}); wanted != *claims {
		t.Fatalf("Wanted:\n%# v\n\nFound:\n%# v", wanted, claims)
	}
}

func parseToken(token, key string) (*jwt.StandardClaims, error) {
	tok, err := jwt.ParseWithClaims(
		token,
		&jwt.StandardClaims{},
		func(*jwt.Token) (interface{}, error) { return []byte(key), nil },
	)
	if err != nil {
		return nil, fmt.Errorf("parsing token: %w", err)
	}

	if tok.Method != jwt.SigningMethodHS512 {
		return nil, fmt.Errorf(
			"wanted method '%v'; found '%v'",
			jwt.SigningMethodHS512,
			tok.Method,
		)
	}

	if c, ok := tok.Claims.(*jwt.StandardClaims); ok {
		return c, nil
	}

	return nil, fmt.Errorf("invalid claims type: %T", tok.Claims)
}

type resetTokenStoreMock struct {
	create func(*ResetToken) error
	get    func(UserID) (*ResetToken, error)
}

func (rtsm *resetTokenStoreMock) Create(rt *ResetToken) error {
	if rtsm.create == nil {
		panic("resetTokenStoreMock: missing `create` hook")
	}
	return rtsm.create(rt)
}

func (rtsm *resetTokenStoreMock) Get(user UserID) (*ResetToken, error) {
	if rtsm.get == nil {
		panic("resetTokenStoreMock: missing `get` hook")
	}
	return rtsm.get(user)
}

func TestBeginRegistration(t *testing.T) {
	now := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	var creds Credentials
	var token ResetToken
	authService := AuthService{
		Creds: &credStoreMock{
			create: func(c *Credentials) error { creds = *c; return nil },
		},
		ResetTokens: &resetTokenStoreMock{
			create: func(rt *ResetToken) error { token = *rt; return nil },
		},
		Notifications: &notificationServiceMock{
			notify: func(UserID, uuid.UUID) error { return nil },
		},
		Hostname:           "auth.example.org",
		TimeFunc:           func() time.Time { return now },
		ResetTokenValidity: 24 * time.Hour,
	}

	if err := authService.BeginRegistration("user"); err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}

	if creds.Username != "user" {
		t.Fatalf(
			"Credentials.Username: wanted 'user'; found '%s'",
			creds.Username,
		)
	}

	if token.User != "user" {
		t.Fatalf(
			"ResetToken.User: wanted 'user'; found '%s'",
			token.User,
		)
	}

	if wanted := now.Add(24 * time.Hour); token.ExpiresAt != wanted {
		t.Fatalf(
			"ResetToken.ExpiresAt: wanted '%s'; found '%s'",
			wanted,
			token.ExpiresAt,
		)
	}
}

func TestBeginRegistration_UserExists(t *testing.T) {
	now := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	var token ResetToken
	authService := AuthService{
		Creds: &credStoreMock{
			create: func(c *Credentials) error { return ErrUserExists },
		},
		ResetTokens: &resetTokenStoreMock{
			create: func(rt *ResetToken) error { token = *rt; return nil },
		},
		Notifications: &notificationServiceMock{
			notify: func(UserID, uuid.UUID) error { return nil },
		},
		Hostname:           "auth.example.org",
		TimeFunc:           func() time.Time { return now },
		ResetTokenValidity: 24 * time.Hour,
	}

	if err := authService.BeginRegistration("user"); err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}

	if token.User != "user" {
		t.Fatalf(
			"ResetToken.User: wanted 'user'; found '%s'",
			token.User,
		)
	}

	if wanted := now.Add(24 * time.Hour); token.ExpiresAt != wanted {
		t.Fatalf(
			"ResetToken.ExpiresAt: wanted '%s'; found '%s'",
			wanted,
			token.ExpiresAt,
		)
	}
}

func TestUpdatePassword(t *testing.T) {
	now := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	tok := uuid.New()
	var creds *Credentials
	authService := AuthService{
		Creds: &credStoreMock{
			update: func(c *Credentials) error { creds = c; return nil },
		},
		ResetTokens: &resetTokenStoreMock{
			get: func(user UserID) (*ResetToken, error) {
				if user == "user" {
					return &ResetToken{
						User:      user,
						Token:     tok,
						ExpiresAt: now.Add(1 * time.Hour),
					}, nil
				}
				return nil, ErrResetTokenNotFound
			},
		},
		TimeFunc: func() time.Time { return now },
	}

	if err := authService.UpdatePassword(&UpdatePassword{
		Credentials: Credentials{
			Username: "user",
			Password: "my-new-password",
		},
		Token: tok,
	}); err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}

	if creds == nil {
		t.Fatal("CredStore: Failed to update credentials")
	}
	if creds.Username != "user" {
		t.Fatalf(
			"Credentials.Username: wanted 'user'; found '%s'",
			creds.Username,
		)
	}
	if creds.Password != "my-new-password" {
		t.Fatalf(
			"Credentials.Password: wanted 'my-new-password'; found '%s'",
			creds.Password,
		)
	}
}
