package types

import (
	"crypto/ecdsa"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type TokenFactory struct {
	Issuer        string
	Audience      string
	TokenValidity time.Duration
	SigningKey    *ecdsa.PrivateKey
}

func (tf *TokenFactory) Create(now time.Time, subject string) (*Token, error) {
	expires := now.Add(tf.TokenValidity)
	token := jwt.NewWithClaims(
		jwt.SigningMethodES512,
		jwt.StandardClaims{
			Subject:   subject,
			Audience:  tf.Audience,
			Issuer:    tf.Issuer,
			IssuedAt:  now.Unix(),
			ExpiresAt: expires.Unix(),
			NotBefore: now.Unix(),
		},
	)
	t, err := token.SignedString(tf.SigningKey)
	if err != nil {
		return nil, fmt.Errorf("signing token: %w", err)
	}
	return &Token{Token: t, Expires: expires}, nil
}
