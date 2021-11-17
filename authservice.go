package main

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"net/mail"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type UserID string

var (
	ErrCredentials       = errors.New("invalid username or password")
	ErrUserExists        = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidResetToken = errors.New("reset token invalid")
	ErrInvalidEmail      = errors.New("invalid email address")
)

type Credentials struct {
	User     UserID `json:"user"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type NotificationType string

const (
	NotificationTypeRegister       NotificationType = "REGISTER"
	NotificationTypeForgotPassword NotificationType = "FORGOT_PASSWORD"
)

type Notification struct {
	Type  NotificationType
	User  UserID
	Email string
	Token string
}

type NotificationService interface {
	Notify(*Notification) error
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

type AuthService struct {
	Creds         CredStore
	Notifications NotificationService
	ResetTokens   ResetTokenFactory
	TokenDetails  TokenDetailsFactory
	TimeFunc      func() time.Time
}

func (as *AuthService) Login(c *Credentials) (*TokenDetails, error) {
	if err := as.Creds.Validate(c); err != nil {
		return nil, fmt.Errorf("validating credentials: %w", err)
	}

	tokenDetails, err := as.TokenDetails.Create(string(c.User))
	if err != nil {
		return nil, fmt.Errorf("creating token details: %w", err)
	}

	return tokenDetails, nil
}

func (as *AuthService) Refresh(refreshToken string) (string, error) {
	var claims jwt.StandardClaims
	if _, err := jwt.ParseWithClaims(
		refreshToken,
		&claims,
		func(*jwt.Token) (interface{}, error) {
			return &as.TokenDetails.RefreshTokens.SigningKey.PublicKey, nil
		},
	); err != nil {
		return "", fmt.Errorf("parsing refresh token: %w", err)
	}

	if err := claims.Valid(); err != nil {
		return "", fmt.Errorf("validating refresh token: %w", err)
	}

	return as.TokenDetails.AccessToken(claims.Subject)
}

func (as *AuthService) Register(user UserID, email string) error {
	parser := mail.AddressParser{}
	if _, err := parser.Parse(email); err != nil {
		return fmt.Errorf("registering user: %w", ErrInvalidEmail)
	}

	if _, err := as.Creds.Users.Get(user); err != nil {
		if !errors.Is(err, ErrUserNotFound) {
			return fmt.Errorf("registering user: %w", err)
		}
	} else {
		// if the error is nil, it means the user was found--return
		// `ErrUserExists`.
		return fmt.Errorf("registering user: %w", ErrUserExists)
	}

	// TODO: Error if user or email already exists
	token, err := as.ResetTokens.Create(as.TimeFunc(), user, email)
	if err != nil {
		return fmt.Errorf("registering user: %w", err)
	}

	if err := as.Notifications.Notify(&Notification{
		Type:  NotificationTypeRegister,
		User:  user,
		Email: email,
		Token: token,
	}); err != nil {
		return fmt.Errorf("notifying registration reset token: %w", err)
	}

	return nil
}

func (as *AuthService) ForgotPassword(user UserID) error {
	u, err := as.Creds.Users.Get(user)
	if err != nil {
		return fmt.Errorf("fetching user: %w", err)
	}

	token, err := as.ResetTokens.Create(as.TimeFunc(), user, u.Email)
	if err != nil {
		return fmt.Errorf("preparing forgot-password notification: %w", err)
	}

	if err := as.Notifications.Notify(&Notification{
		Type:  NotificationTypeForgotPassword,
		User:  user,
		Email: u.Email,
		Token: token,
	}); err != nil {
		return fmt.Errorf("notifying forgot-password reset token: %w", err)
	}

	return nil
}

type UpdatePassword struct {
	User     UserID `json:"user"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

func (as *AuthService) UpdatePassword(up *UpdatePassword) error {
	claims, err := as.ResetTokens.Claims(up.Token)
	if err != nil {
		return fmt.Errorf("updating password: %w", err)
	}

	// We deliberately want to return `ErrResetTokenNotFound` in this case so
	// as not to give attackers unnecessary information. See OWASP link above.
	if err := claims.Valid(); err != nil {
		return fmt.Errorf("updating password: %w", ErrInvalidResetToken)
	}

	if err := as.Creds.Upsert(&Credentials{
		User:     up.User,
		Email:    claims.Email,
		Password: up.Password,
	}); err != nil {
		return fmt.Errorf("updating password: %w", err)
	}

	return nil
}
