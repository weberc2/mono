package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
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
	hashed, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
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
				WildcardAudience: "*.example.com",
				TokenValidity:    15 * time.Minute,
				SigningKey:       accessKey,
				SigningMethod:    jwt.SigningMethodES512,
			},
			RefreshTokens: TokenFactory{
				Issuer:           "issuer",
				WildcardAudience: "*.example.com",
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
		Audience:  "*.example.com",
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
		Audience:  "*.example.com",
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

type notificationServiceMock struct {
	notify func(UserID, uuid.UUID) error
}

func (nsm *notificationServiceMock) Notify(u UserID, t uuid.UUID) error {
	if nsm.notify == nil {
		panic("notificationServiceMock: missing `notify` hook")
	}
	return nsm.notify(u, t)
}

func TestRegister(t *testing.T) {
	now := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	var entry UserEntry
	var token ResetToken
	authService := AuthService{
		Creds: CredStore{&userStoreMock{
			create: func(e *UserEntry) error { entry = *e; return nil },
		}},
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

	if err := authService.Register("user"); err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}

	if entry.User != "user" {
		t.Fatalf(
			"UserEntry.User: wanted 'user'; found '%s'",
			entry.User,
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
		Creds: CredStore{&userStoreMock{
			create: func(*UserEntry) error { return ErrUserExists },
		}},
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

	if err := authService.Register("user"); err != nil {
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
	const password = "osakldflhkjewadfkjsfduIHUHKJGFU"
	hashed, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}
	var entry *UserEntry
	authService := AuthService{
		Creds: CredStore{&userStoreMock{
			update: func(e *UserEntry) error { entry = e; return nil },
		}},
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
		User:     "user",
		Password: password,
		Token:    tok,
	}); err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}

	if entry == nil {
		t.Fatal("CredStore: Failed to update credentials")
	}
	if entry.User != "user" {
		t.Fatalf(
			"UserEntry.User: wanted 'user'; found '%s'",
			entry.User,
		)
	}
	if bytes.Equal(entry.PasswordHash, hashed) {
		t.Fatalf(
			"UserEntry.PasswordHash: wanted '%s'; found '%s'",
			base64.RawStdEncoding.EncodeToString(hashed),
			base64.RawStdEncoding.EncodeToString(entry.PasswordHash),
		)
	}
}
