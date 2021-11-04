package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	pz "github.com/weberc2/httpeasy"
)

type notificationServiceMock struct {
	notify func(UserID, uuid.UUID) error
}

func (nsm *notificationServiceMock) Notify(u UserID, t uuid.UUID) error {
	if nsm.notify == nil {
		panic("notificationServiceMock: `notify` hook unset")
	}
	return nsm.notify(u, t)
}

func main() {
	const (
		issuer            = "weberc2.com"
		wildcardAudience  = "*.weberc2.com"
		accessSigningKey  = "access-signing-key"
		refreshSigningKey = "refresh-signing-key"
	)
	authService := AuthHTTPService{AuthService{
		Creds:         &MemCredStore{},
		ResetTokens:   &MemResetTokenStore{},
		Notifications: ConsoleNotificationService{},
		Hostname:      "auth.weberc2.com",
		TokenDetails: TokenDetailsFactory{
			AccessTokens: TokenFactory{
				Issuer:           issuer,
				WildcardAudience: wildcardAudience,
				TokenValidity:    15 * time.Minute,
				SigningKey:       []byte(accessSigningKey),
				SigningMethod:    jwt.SigningMethodHS512,
			},
			RefreshTokens: TokenFactory{
				Issuer:           issuer,
				WildcardAudience: wildcardAudience,
				TokenValidity:    7 * 24 * time.Hour,
				SigningKey:       []byte(refreshSigningKey),
				SigningMethod:    jwt.SigningMethodHS512,
			},
			TimeFunc: time.Now,
		},
		ResetTokenValidity: 1 * time.Hour,
		TimeFunc:           time.Now,
	}}

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}

	log.Printf("Listening on %s", addr)
	if err := http.ListenAndServe(
		addr,
		pz.Register(pz.JSONLog(os.Stderr), authService.Routes()...),
	); err != nil {
		log.Fatal(err)
	}
}
