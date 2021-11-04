package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	pz "github.com/weberc2/httpeasy"
)

func main() {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}

	hostName := os.Getenv("HOST_NAME")
	if hostName == "" {
		hostName = addr
	}

	issuer := os.Getenv("ISSUER")
	if issuer == "" {
		log.Fatal("Missing required env var: ISSUER")
	}

	audience := os.Getenv("AUDIENCE")
	if audience == "" {
		log.Fatal("Missing required env var: AUDIENCE")
	}

	accessSigningKey := os.Getenv("ACCESS_KEY")
	if accessSigningKey == "" {
		log.Fatal("Missing required env var: ACCESS_KEY")
	}

	refreshSigningKey := os.Getenv("REFRESH_KEY")
	if refreshSigningKey == "" {
		log.Fatal("Missing required env var: REFRESH_KEY")
	}

	authService := AuthHTTPService{AuthService{
		Creds:         &MemCredStore{},
		ResetTokens:   &MemResetTokenStore{},
		Notifications: ConsoleNotificationService{},
		Hostname:      hostName,
		TokenDetails: TokenDetailsFactory{
			AccessTokens: TokenFactory{
				Issuer:           issuer,
				WildcardAudience: audience,
				TokenValidity:    15 * time.Minute,
				SigningKey:       []byte(accessSigningKey),
				SigningMethod:    jwt.SigningMethodHS512,
			},
			RefreshTokens: TokenFactory{
				Issuer:           issuer,
				WildcardAudience: audience,
				TokenValidity:    7 * 24 * time.Hour,
				SigningKey:       []byte(refreshSigningKey),
				SigningMethod:    jwt.SigningMethodHS512,
			},
			TimeFunc: time.Now,
		},
		ResetTokenValidity: 1 * time.Hour,
		TimeFunc:           time.Now,
	}}

	log.Printf("Listening on %s", addr)
	if err := http.ListenAndServe(
		addr,
		pz.Register(pz.JSONLog(os.Stderr), authService.Routes()...),
	); err != nil {
		log.Fatal(err)
	}
}
