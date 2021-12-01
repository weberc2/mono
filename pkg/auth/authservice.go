package auth

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"log"
	"net/mail"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/weberc2/auth/pkg/types"
	pz "github.com/weberc2/httpeasy"
)

var (
	ErrCredentials = &pz.HTTPError{
		Status:  401,
		Message: "invalid username or password",
	}
	ErrUnauthorized      = &pz.HTTPError{Status: 401, Message: "Unauthorized"}
	ErrUserExists        = errors.New("user already exists")
	ErrInvalidResetToken = errors.New("reset token invalid")
	ErrInvalidEmail      = errors.New("invalid email address")
)

type TokenFactory struct {
	Issuer        string
	Audience      string
	TokenValidity time.Duration
	SigningKey    *ecdsa.PrivateKey
}

func (tf *TokenFactory) Create(now time.Time, subject string) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodES512,
		jwt.StandardClaims{
			Subject:   subject,
			Audience:  tf.Audience,
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
	Notifications types.NotificationService
	ResetTokens   ResetTokenFactory
	TokenDetails  TokenDetailsFactory
	Codes         TokenFactory
	TimeFunc      func() time.Time
}

func (as *AuthService) Login(c *types.Credentials) (*TokenDetails, error) {
	if err := as.Creds.Validate(c); err != nil {
		return nil, fmt.Errorf("validating credentials: %w", err)
	}

	tokenDetails, err := as.TokenDetails.Create(string(c.User))
	if err != nil {
		return nil, fmt.Errorf("creating token details: %w", err)
	}

	return tokenDetails, nil
}

func (as *AuthService) LoginAuthCode(c *types.Credentials) (string, error) {
	if err := as.Creds.Validate(c); err != nil {
		return "", fmt.Errorf("validating credentials: %w", err)
	}

	code, err := as.Codes.Create(as.TimeFunc(), string(c.User))
	if err != nil {
		return "", fmt.Errorf("creating auth code: %w", err)
	}

	return code, nil
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

func (as *AuthService) Register(user types.UserID, email string) error {
	parser := mail.AddressParser{}
	if _, err := parser.Parse(email); err != nil {
		return fmt.Errorf("registering user: %w", ErrInvalidEmail)
	}

	if _, err := as.Creds.Users.Get(user); err != nil {
		if !errors.Is(err, types.ErrUserNotFound) {
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

	if err := as.Notifications.Notify(&types.Notification{
		Type:  types.NotificationTypeRegister,
		User:  user,
		Email: email,
		Token: token,
	}); err != nil {
		return fmt.Errorf("notifying registration reset token: %w", err)
	}

	return nil
}

func (as *AuthService) ForgotPassword(user types.UserID) error {
	u, err := as.Creds.Users.Get(user)
	if err != nil {
		return fmt.Errorf("fetching user: %w", err)
	}

	token, err := as.ResetTokens.Create(as.TimeFunc(), user, u.Email)
	if err != nil {
		return fmt.Errorf("preparing forgot-password notification: %w", err)
	}

	if err := as.Notifications.Notify(&types.Notification{
		Type:  types.NotificationTypeForgotPassword,
		User:  user,
		Email: u.Email,
		Token: token,
	}); err != nil {
		return fmt.Errorf("notifying forgot-password reset token: %w", err)
	}

	return nil
}

type UpdatePassword struct {
	User     types.UserID `json:"user"`
	Password string       `json:"password"`
	Token    string       `json:"token"`
}

func (as *AuthService) UpdatePassword(up *UpdatePassword) error {
	claims, err := as.ResetTokens.Claims(up.Token)
	if err != nil {
		return fmt.Errorf("updating password: %w", err)
	}

	// We deliberately want to return `ErrInvalidResetToken` in this case so
	// as not to give attackers unnecessary information. See OWASP link above.
	if err := claims.Valid(); err != nil {
		return fmt.Errorf("updating password: %w", ErrInvalidResetToken)
	}

	if err := as.Creds.Upsert(&types.Credentials{
		User:     up.User,
		Email:    claims.Email,
		Password: up.Password,
	}); err != nil {
		return fmt.Errorf("updating password: %w", err)
	}

	return nil
}

func (as *AuthService) Exchange(code string) (*TokenDetails, error) {
	var claims Claims
	if _, err := jwt.ParseWithClaims(
		code,
		&claims,
		func(*jwt.Token) (interface{}, error) {
			return &as.Codes.SigningKey.PublicKey, nil
		},
	); err != nil {
		log.Printf("jwt.ParseWithClaims(): %v", err)
		return nil, ErrUnauthorized
	}

	if err := claims.Valid(); err != nil {
		log.Printf("Claims.Valid(): %v", err)
		return nil, ErrUnauthorized
	}

	tokens, err := as.TokenDetails.Create(claims.Subject)
	if err != nil {
		return nil, fmt.Errorf("creating access and refresh tokens: %w", err)
	}

	return tokens, nil
}
