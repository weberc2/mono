package main

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/google/uuid"
	pz "github.com/weberc2/httpeasy"
	"github.com/weberc2/mono/pkg/auth/client"
	"github.com/weberc2/mono/pkg/comments"
	"github.com/weberc2/mono/pkg/comments/types"
	"github.com/weberc2/mono/pkg/pgcommentsstore"
)

func main() {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}

	loginURL := os.Getenv("LOGIN_URL")
	if loginURL == "" {
		log.Fatal("missing required env var: LOGIN_URL")
	}

	registerURL := os.Getenv("REGISTER_URL")
	if registerURL == "" {
		log.Fatal("missing required env var: REGISTER_URL")
	}

	authBaseURL := os.Getenv("AUTH_BASE_URL")
	if authBaseURL == "" {
		log.Fatal("missing required env var: AUTH_BASE_URL")
	}

	baseURLString := os.Getenv("BASE_URL")
	if baseURLString == "" {
		log.Fatal("missing required env var: BASE_URL")
	}
	baseURL, err := url.Parse(baseURLString)
	if err != nil {
		log.Fatalf("error parsing `BASE_URL` env var: %v", err)
	}

	cookieEncryptionKey := os.Getenv("COOKIE_ENCRYPTION_KEY")
	if cookieEncryptionKey == "" {
		log.Fatal("missing required env var: COOKIE_ENCRYPTION_KEY")
	}

	accessTokenPublicKey := os.Getenv("ACCESS_KEY")
	if accessTokenPublicKey == "" {
		log.Fatal("missing required env var: ACCESS_KEY")
	}
	key, err := decodeKey(accessTokenPublicKey)
	if err != nil {
		log.Fatalf("decoding ACCESS_KEY: %v", err)
	}

	commentsStore, err := pgcommentsstore.OpenEnv()
	if err != nil {
		log.Fatalf("creating postgres comments store client: %v", err)
	}

	if err := commentsStore.EnsureTable(); err != nil {
		log.Fatalf("ensuring comments table exists: %v", err)
	}

	authenticator := client.Authenticator{Key: key}

	commentsService := comments.AuthCommentsService{
		CommentsService: comments.CommentsService{
			Comments: comments.CommentsModel{
				CommentsStore: commentsStore,
				IDFunc: func() types.CommentID {
					return types.CommentID(uuid.NewString())
				},
				TimeFunc: time.Now,
			},
		},
		Auth: comments.Auth{
			AuthType:      client.AuthTypeClientProgram{},
			Authenticator: authenticator,
		},
	}

	webServerAuth := client.AuthTypeWebServer{
		WebServerApp: client.WebServerApp{
			Client:          client.DefaultClient(authBaseURL),
			BaseURL:         baseURL,
			DefaultRedirect: "/",
			Key:             cookieEncryptionKey,
		},
	}

	webServer := comments.AuthWebServer{
		WebServer: comments.WebServer{
			LoginURL:         loginURL,
			RegisterURL:      registerURL,
			LogoutPath:       "/auth/logout",
			BaseURL:          baseURLString,
			Comments:         commentsService.Comments,
			AuthCallbackPath: "/auth/callback",
		},
		Auth: comments.Auth{
			AuthType:      &webServerAuth,
			Authenticator: authenticator,
		},
	}

	if err := http.ListenAndServe(addr, pz.Register(
		pz.JSONLog(os.Stderr),
		append(
			append(
				webServer.Routes(),
				webServerAuth.AuthCodeCallbackRoute(webServer.AuthCallbackPath),
				webServerAuth.LogoutRoute(webServer.LogoutPath),
			),
			commentsService.Routes()...,
		)...,
	)); err != nil {
		log.Fatal(err)
	}
}

func decodeKey(encoded string) (*ecdsa.PublicKey, error) {
	data := []byte(encoded)
	for {
		block, rest := pem.Decode(data)
		if block.Type == "PUBLIC KEY" {
			pub, err := x509.ParsePKIXPublicKey(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("parsing x509 PKIX public key: %v", err)
			}
			if key, ok := pub.(*ecdsa.PublicKey); ok {
				return key, nil
			}
			return nil, fmt.Errorf(
				"invalid key type; wanted *ecdsa.PublicKey; found %T",
				pub,
			)
		}
		if len(rest) < 1 {
			return nil, io.EOF
		}
		data = rest
	}
}
