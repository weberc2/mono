package main

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/dgrijalva/jwt-go"
	pz "github.com/weberc2/httpeasy"
)

type PostID string
type CommentID string
type UserID string

type Comment struct {
	ID       CommentID
	Parent   CommentID
	Author   UserID
	Created  time.Time
	Modified time.Time
	Body     string
}

type noopPostStore struct{}

func (nps noopPostStore) Exists(PostID) error { return nil }

func main() {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}

	bucket := os.Getenv("BUCKET")
	if bucket == "" {
		log.Fatal("Missing required env var: BUCKET")
	}

	accessTokenPublicKey := os.Getenv("ACCESS_KEY")
	if accessTokenPublicKey == "" {
		log.Fatal("Missing required env var: ACCESS_KEY")
	}
	key, err := decodeKey(accessTokenPublicKey)
	if err != nil {
		log.Fatalf("decoding ACCESS_KEY: %v", err)
	}

	commentsService := CommentsService{
		Store: CommentStore{
			Bucket:      bucket,
			Prefix:      "",
			ObjectStore: &S3ObjectStore{s3.New(session.New())},
			PostStore:   noopPostStore{},
		},
	}

	http.ListenAndServe(addr, pz.Register(
		pz.JSONLog(os.Stderr),
		pz.Route{
			Method:  "GET",
			Path:    "/posts/{post-id}/comments/{comment-id}/comments",
			Handler: commentsService.PostComments,
		},
		pz.Route{
			Method:  "POST",
			Path:    "/posts/{post-id}/comments",
			Handler: auth(key, commentsService.PutComment),
		},
		pz.Route{
			Method:  "GET",
			Path:    "/posts/{post-id}/comments/{comment-id}",
			Handler: commentsService.GetComment,
		},
	))
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

func auth(key *ecdsa.PublicKey, handler pz.Handler) pz.Handler {
	return func(r pz.Request) pz.Response {
		authorization := r.Headers.Get("Authorization")
		if !strings.HasPrefix(authorization, "Bearer ") {
			return pz.Unauthorized(nil, struct {
				Message, Error string
			}{
				Message: "invalid 'Authorization' header",
				Error:   "missing 'Bearer' prefix",
			})
		}

		var claims jwt.StandardClaims
		if _, err := jwt.ParseWithClaims(
			authorization[len("Bearer "):],
			&claims,
			func(*jwt.Token) (interface{}, error) {
				return key, nil
			},
		); err != nil {
			return pz.Unauthorized(nil, struct {
				Message, Error string
			}{
				Message: "invalid 'Authorization' header",
				Error:   err.Error(),
			})
		}

		if err := claims.Valid(); err != nil {
			return pz.Unauthorized(nil, struct {
				Message, Error string
			}{
				Message: "invalid access token claim(s)",
				Error:   err.Error(),
			})
		}

		r.Headers.Set("User", claims.Subject)
		return handler(r)
	}
}
