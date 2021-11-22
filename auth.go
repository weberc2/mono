package main

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/weberc2/auth/pkg/client"
	pz "github.com/weberc2/httpeasy"
)

type AuthType interface {
	Validate(key *ecdsa.PublicKey, r pz.Request) (string, *AuthErr)
}

type AuthTypeClientProgram struct{}

func (atcp AuthTypeClientProgram) Validate(
	key *ecdsa.PublicKey,
	r pz.Request,
) (string, *AuthErr) {
	authorization := r.Headers.Get("Authorization")
	if !strings.HasPrefix(authorization, "Bearer ") {
		return "", &AuthErr{
			"invalid access token",
			fmt.Errorf("missing `Bearer` prefix"),
		}
	}

	subject, err := validateAccessToken(authorization[len("Bearer "):], key)
	if err != nil {
		return "", &AuthErr{"invalid access token", err}
	}

	return subject, nil
}

type AuthTypeWebServer struct {
	Auth client.Client
}

func (atws *AuthTypeWebServer) Validate(
	key *ecdsa.PublicKey,
	r pz.Request,
) (string, *AuthErr) {
	accessCookie, err := r.Cookie("Access-Token")
	if err != nil {
		return "", &AuthErr{"missing `Access-Token` cookie", err}
	}

	refreshCookie, err := r.Cookie("Refresh-Token")
	if err != nil {
		return "", &AuthErr{"missing `Refresh-Token` cookie", err}
	}

	subject, err := validateAccessToken(accessCookie.Value, key)
	if err != nil {
		if err, ok := err.(*jwt.ValidationError); ok {
			masked := err.Errors & jwt.ValidationErrorExpired
			if masked == jwt.ValidationErrorExpired {
				tokens, err := atws.Auth.Refresh(refreshCookie.Value)
				if err != nil {
					return "", &AuthErr{"refreshing access token", err}
				}
				accessCookie.Value = tokens.AccessToken
				refreshCookie.Value = tokens.RefreshToken
				return subject, nil
			}
			return "", &AuthErr{"validating access token", err}
		}
		return "", &AuthErr{"parsing access token", err}
	}

	return subject, nil
}

func validateAccessToken(token string, key *ecdsa.PublicKey) (string, error) {
	var claims jwt.StandardClaims
	if _, err := jwt.ParseWithClaims(
		token,
		&claims,
		func(*jwt.Token) (interface{}, error) {
			return key, nil
		},
	); err != nil {
		return "", err
	}
	return claims.Subject, nil
}

type AuthErr struct {
	Message string `json:"message"`
	Error   error  `json:"error"`
}

func (err *AuthErr) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Message string `json:"message"`
		Error   string `json:"error"`
	}{
		err.Message,
		err.Error.Error(),
	})
}

type Authenticator struct {
	Auth client.Client
	Key  *ecdsa.PublicKey
}

func (a *Authenticator) AuthN(authType AuthType, h pz.Handler) pz.Handler {
	return func(r pz.Request) pz.Response {
		subject, err := authType.Validate(a.Key, r)
		var message string
		if err != nil {
			message = "authentication error"
		} else {
			message = "authentication successful"
		}

		r.Headers.Add("User", subject)
		return h(r).WithLogging(struct {
			Message string   `json:"message"`
			User    string   `json:"user,omitempty"`
			Error   *AuthErr `json:"error,omitempty"`
		}{
			Message: message,
			User:    subject,
			Error:   err,
		})
	}
}

func (a *Authenticator) AuthZ(authType AuthType, h pz.Handler) pz.Handler {
	return func(r pz.Request) pz.Response {
		subject, err := authType.Validate(a.Key, r)
		if err != nil {
			return pz.Unauthorized(nil, err)
		}
		return h(r).
			WithHeaders(http.Header{"User": []string{subject}}).
			WithLogging(struct {
				Message string `json:"message"`
				User    string `json:"user"`
			}{
				Message: "authentication successful",
				User:    subject,
			})
	}
}
