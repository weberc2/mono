package main

import (
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

type UserID string

var ErrCredentials = errors.New("Invalid username or password")
var ErrUserExists = errors.New("User already exists")
var ErrUserNotFound = errors.New("User not found")

type Credentials struct {
	User     UserID `json:"user"`
	Password string `json:"password"`
}

type CredStore interface {
	Validate(*Credentials) error
	Update(*Credentials) error
	Create(*Credentials) error
}

type NotificationService interface {
	Notify(user UserID, resetToken uuid.UUID) error
}

var ErrResetTokenNotFound = errors.New("Reset token not found or expired")

type ResetToken struct {
	User      UserID
	Token     uuid.UUID
	ExpiresAt time.Time
}

type ResetTokenStore interface {
	// Creates or updates the ResetToken for a given user.
	Create(*ResetToken) error

	// Retrieves the ResetToken for a given user.
	Get(UserID) (*ResetToken, error)
}

type TokenFactory struct {
	Issuer           string
	WildcardAudience string
	TokenValidity    time.Duration
	SigningKey       *ecdsa.PrivateKey
	SigningMethod    jwt.SigningMethod
}

func (tf *TokenFactory) Create(now time.Time, subject string) (string, error) {
	token := jwt.NewWithClaims(
		tf.SigningMethod,
		jwt.StandardClaims{
			Subject:   subject,
			Audience:  tf.WildcardAudience,
			Issuer:    tf.Issuer,
			IssuedAt:  now.Unix(),
			ExpiresAt: now.Add(tf.TokenValidity).Unix(),
			NotBefore: now.Unix(),
		},
	)
	return token.SignedString(tf.SigningKey)
}

type TokenDetails struct {
	AccessToken  string
	RefreshToken string
}

type TokenDetailsFactory struct {
	AccessTokens  TokenFactory
	RefreshTokens TokenFactory
	TimeFunc      func() time.Time
}

func (tdf *TokenDetailsFactory) Create(subject string) (*TokenDetails, error) {
	now := tdf.TimeFunc()
	accessToken, err := tdf.AccessTokens.Create(now, subject)
	if err != nil {
		return nil, fmt.Errorf("creating access token: %w", err)
	}

	refreshToken, err := tdf.RefreshTokens.Create(now, subject)
	if err != nil {
		return nil, fmt.Errorf("creating refresh token: %w", err)
	}

	return &TokenDetails{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (tdf *TokenDetailsFactory) AccessToken(subject string) (string, error) {
	return tdf.AccessTokens.Create(tdf.TimeFunc(), subject)
}

type AuthService struct {
	Creds              CredStore
	ResetTokens        ResetTokenStore
	Notifications      NotificationService
	Hostname           string
	ResetTokenValidity time.Duration
	TokenDetails       TokenDetailsFactory
	TimeFunc           func() time.Time
}

func (as *AuthService) Login(c *Credentials) (*TokenDetails, error) {
	if err := as.Creds.Validate(c); err != nil {
		return nil, err
	}

	return as.TokenDetails.Create(string(c.User))
}

func (as *AuthService) Refresh(refreshToken string) (string, error) {
	var claims jwt.StandardClaims
	if _, err := jwt.ParseWithClaims(
		refreshToken,
		&claims,
		func(*jwt.Token) (interface{}, error) {
			return as.TokenDetails.RefreshTokens.SigningKey, nil
		},
	); err != nil {
		return "", fmt.Errorf("parsing refresh token: %w", err)
	}

	if err := claims.Valid(); err != nil {
		return "", fmt.Errorf("validating refresh token: %w", err)
	}

	return as.TokenDetails.AccessToken(claims.Subject)
}

func (as *AuthService) Register(user UserID) error {
	if err := as.Creds.Create(&Credentials{
		User:     user,
		Password: uuid.NewString(),
	}); err != nil {
		if !errors.Is(err, ErrUserExists) {
			return fmt.Errorf("Beginning registration: %w", err)
		}
		// If the user already exists, we'll continue.
	}

	token := ResetToken{
		User:      user,
		Token:     uuid.New(),
		ExpiresAt: as.TimeFunc().UTC().Add(as.ResetTokenValidity),
	}

	// Create a new token, possibly overwriting an existing token if the user
	// already existed.
	if err := as.ResetTokens.Create(&token); err != nil {
		return fmt.Errorf("Beginning registration: %w", err)
	}

	if err := as.Notifications.Notify(
		user,
		token.Token,
	); err != nil {
		return fmt.Errorf("Beginning registration: %w", err)
	}

	return nil
}

type UpdatePassword struct {
	User     UserID
	Password string
	Token    uuid.UUID
}

func (up *UpdatePassword) UnmarshalJSON(data []byte) error {
	var payload struct {
		User     UserID `json:"user"`
		Password string `json:"password"`
		Token    string `json:"token"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}
	token, err := uuid.Parse(payload.Token)
	if err != nil {
		return fmt.Errorf("parsing token into UUID: %w", err)
	}
	up.User = payload.User
	up.Password = payload.Password
	up.Token = token
	return nil
}

func (as *AuthService) UpdatePassword(up *UpdatePassword) error {
	// https://cheatsheetseries.owasp.org/cheatsheets/Forgot_Password_Cheat_Sheet.html
	now := as.TimeFunc()

	resetToken, err := as.ResetTokens.Get(up.User)
	if err != nil {
		return fmt.Errorf("updating password: %w", err)
	}

	// We deliberately want to return `ErrResetTokenNotFound` in these cases so
	// as to not give attackers unnecessary information. See owasp link above.
	if resetToken.Token != up.Token || resetToken.ExpiresAt.Before(now) {
		log.Printf("resetToken.ExpiresAt.Before(now): %v", resetToken.ExpiresAt.Before(now))
		return fmt.Errorf("updating password: %w", ErrResetTokenNotFound)
	}

	if err := as.Creds.Update(&Credentials{
		User:     up.User,
		Password: up.Password,
	}); err != nil {
		return fmt.Errorf("updating password: %w", err)
	}

	return nil
}
