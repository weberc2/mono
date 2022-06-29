package types

import (
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	pz "github.com/weberc2/httpeasy"
)

type Claims struct {
	User  UserID
	Email string
	jwt.StandardClaims
}

type ResetTokenFactory TokenFactory

func (rtf *ResetTokenFactory) Create(
	now time.Time,
	user UserID,
	email string,
) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodES512,
		Claims{
			User:  user,
			Email: email,
			StandardClaims: jwt.StandardClaims{
				Subject:   string(user),
				Audience:  rtf.Audience,
				Issuer:    rtf.Issuer,
				IssuedAt:  now.Unix(),
				ExpiresAt: now.Add(rtf.TokenValidity).Unix(),
				NotBefore: now.Unix(),
			},
		},
	)
	return token.SignedString(rtf.SigningKey)
}

func (rtf *ResetTokenFactory) Claims(token string) (*Claims, error) {
	var claims Claims
	if _, err := jwt.ParseWithClaims(
		token,
		&claims,
		func(*jwt.Token) (interface{}, error) {
			return &rtf.SigningKey.PublicKey, nil
		},
	); err != nil {
		return nil, InvalidRefreshTokenErr(err)
	}
	return &claims, nil
}

func InvalidRefreshTokenErr(err error) *pz.HTTPError {
	return &pz.HTTPError{
		Status:  http.StatusBadRequest,
		Message: "invalid password reset token",
		Cause_:  err,
	}
}
