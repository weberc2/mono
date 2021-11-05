package main

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
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

	accessSigningKeyEncoded := os.Getenv("ACCESS_PRIVATE_KEY")
	if accessSigningKeyEncoded == "" {
		log.Fatal("Missing required env var: ACCESS_PRIVATE_KEY")
	}
	accessSigningKey, err := decodeKey(accessSigningKeyEncoded)
	if err != nil {
		log.Fatalf("Decoding access key: %v", err)
	}

	refreshSigningKeyEncoded := os.Getenv("REFRESH_PRIVATE_KEY")
	if refreshSigningKeyEncoded == "" {
		log.Fatal("Missing required env var: REFRESH_PRIVATE_KEY")
	}
	refreshSigningKey, err := decodeKey(refreshSigningKeyEncoded)
	if err != nil {
		log.Fatalf("Decoding refresh key: %v", err)
	}

	authService := AuthHTTPService{AuthService{
		Creds: CredStore{Users: &DynamoDBUserStore{
			Client: dynamodb.New(session.New()),
			Table:  "Users",
		}},
		ResetTokens:   &MemResetTokenStore{},
		Notifications: ConsoleNotificationService{},
		Hostname:      hostName,
		TokenDetails: TokenDetailsFactory{
			AccessTokens: TokenFactory{
				Issuer:           issuer,
				WildcardAudience: audience,
				TokenValidity:    15 * time.Minute,
				SigningKey:       accessSigningKey,
				SigningMethod:    jwt.SigningMethodES512,
			},
			RefreshTokens: TokenFactory{
				Issuer:           issuer,
				WildcardAudience: audience,
				TokenValidity:    7 * 24 * time.Hour,
				SigningKey:       refreshSigningKey,
				SigningMethod:    jwt.SigningMethodES512,
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

func decodeKey(encoded string) (*ecdsa.PrivateKey, error) {
	data := []byte(encoded)
	for {
		block, rest := pem.Decode(data)
		if block.Type != "PRIVATE KEY" {
			if len(rest) > 0 {
				data = rest
				continue
			}
			return nil, fmt.Errorf("PEM data is missing a 'PRIVATE KEY' block")
		}
		key, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parsing ecdsa private key: %w", err)
		}
		return key, nil
	}
}
