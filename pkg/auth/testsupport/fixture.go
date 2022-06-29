package testsupport

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/weberc2/mono/pkg/auth/types"
	"golang.org/x/crypto/bcrypt"

	. "github.com/weberc2/mono/pkg/prelude"
)

func NowTimeFunc() time.Time { return Now }

var (
	GoodPasswordHash = Must(HashBcrypt(GoodPassword))

	Now = time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)

	PasswordResetNotification = types.Notification{
		Type:  types.NotificationTypeForgotPassword,
		User:  User,
		Email: Email,
		Token: ResetToken,
	}
	RegistrationNotification = types.Notification{
		Type:  types.NotificationTypeRegister,
		User:  User,
		Email: Email,
		Token: ResetToken,
	}

	ResetToken   = Must(ResetTokenFactory.Create(Now, User, Email))
	AccessToken  = *Must(AccessTokenFactory.Create(Now, string(User)))
	RefreshToken = *Must(RefreshTokenFactory.Create(Now, string(User)))
	AuthCode     = *Must(CodesTokenFactory.Create(Now, string(User)))

	AccessTokenFactory = types.TokenFactory{
		Issuer:        "issuer",
		Audience:      "audience",
		TokenValidity: 15 * time.Minute,
		SigningKey:    AccessSigningKey,
	}
	RefreshTokenFactory = types.TokenFactory{
		Issuer:        "issuer",
		Audience:      "audience",
		TokenValidity: 7 * 24 * time.Hour,
		SigningKey:    RefreshSigningKey,
	}
	ResetTokenFactory = types.ResetTokenFactory{
		Issuer:        "issuer",
		Audience:      "audience",
		TokenValidity: 1 * time.Hour,
		SigningKey:    ResetSigningKey,
	}
	CodesTokenFactory = types.TokenFactory{
		Issuer:        "issuer",
		Audience:      "audience",
		TokenValidity: time.Minute,
		SigningKey:    CodesSigningKey,
	}
	AccessSigningKey  = Must(parseKey(AccessSigningKeyString))
	RefreshSigningKey = Must(parseKey(RefreshSigningKeyString))
	ResetSigningKey   = Must(parseKey(ResetSigningKeyString))
	CodesSigningKey   = Must(parseKey(CodesSigningKeyString))
)

const (
	GoodPassword = ";oasdfipas#@#$OPYODF:;asdf"

	BaseURL                 = "https://auth.example.org"
	RedirectDomain          = "app.example.org"
	DefaultRedirectLocation = "https://" + RedirectDomain + "/default-redirect"

	AccessSigningKeyString = `-----BEGIN PRIVATE KEY-----
MIHcAgEBBEIBb4gjfi9dZnm6jypDJ1/44jUYYPaAizXv7QQPG14aj9W1pwoULDuM
ni71Zi68U8NJhB/dfHgvviK8a8289lysux+gBwYFK4EEACOhgYkDgYYABACD5lbL
9RtF/WKFyUpn8FBJ1QZHvsxcfgpSlvGPyJa3pP9NbofkFL5Xuh9Yd5oFp40xQhJv
f9MBqFs4XHv363V+egB5HQFk0oQeiwl8kNfCgTsZzM4CMytyVQZty2zM9CKXG5m7
EjWmjtDDCSEnLodzVVtL89VNxPI97T4P5QFolAMezg==
-----END PRIVATE KEY-----`
	RefreshSigningKeyString = `-----BEGIN PRIVATE KEY-----
MIHcAgEBBEIANg/VI7PQKRnNeBz4WKfQWFrQUOfuelQeNMTh9ItWpCKqHB5yb5ba
DMJo4lEXjtduf/vvjPNqWurHGuEAW3aM3n+gBwYFK4EEACOhgYkDgYYABAGidC1I
tlhV5Xgs4xb+co5TI2YIA2huX47u18zZNs8wCmGxwPZ6fQlZW5SCekdNS4K6rocr
TkOM9C1EWEA18dyYngDcIurK/D5Pia3FaorX14KMxduUafX/hhOmWChBrIcK3FWW
gpjZ21DFCBpFh83l3tCrfD+yDXElY9EAg8Xur3vSfg==
-----END PRIVATE KEY-----`
	ResetSigningKeyString = `-----BEGIN PRIVATE KEY-----
MIHcAgEBBEIBteoGMRxbAQSI2z9nhD/GBcMVfecuyG58swlqZZDRQ8aUTcmaL371
+9cSBTI6AFNRWl6Fh0/kD4Kyg8UR+4R8fdWgBwYFK4EEACOhgYkDgYYABAEuc5pj
bi3AWn/XJ8xxVn8cDuvnqXEWec+/oiFkJkvlqe0YTA/mz/lmoIgQget6nMVAXUa0
C0Gwvg5hxJ6EF7+ZWwFLFgcyCWW2tezZyNqi7BBW6dAlRGOun6VrldPAJFW96cl8
i5q05kD3gwd3T6OmOv0gCoVYvDhHwZLNuVOUHYVUjg==
-----END PRIVATE KEY-----`
	CodesSigningKeyString = `-----BEGIN PRIVATE KEY-----
MIHcAgEBBEIAPCYJluF6sic9MEGZAl+h3D+heZpBL4+KdBeofuVkjVjA+FYghsPI
7sOsI8t005xekngXMtL6rUlUvDx7wU7WU8+gBwYFK4EEACOhgYkDgYYABAB5BZdD
RrGMdKPeQ7qVOF0Vx8da49z0a49rM18+9lbStPXaLiGmJGNajBrcUSydL6bn52Fw
2fwSJOoPX2blD/ijlAFaKrER8VYzy98B7heWO5RHACE2ZW+DYuBBAMdGXpO+HfJu
zEBS0EsiFH2M/MoLWgvkBmeC+TdCsr761bHQYYVDMw==
-----END PRIVATE KEY-----`

	User  types.UserID = "user"
	Email              = "user@example.org"
)

func HashBcrypt(password string) ([]byte, error) {
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, fmt.Errorf("bcrypt-hashing password: %v", err)
	}
	return hash, nil
}

func parseKey(keyString string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(keyString))

	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parsing x509 EC private key: %v", err)
	}

	return key, nil
}
