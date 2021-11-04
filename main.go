package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
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

func main() {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}

	bucket := os.Getenv("BUCKET")
	if bucket == "" {
		log.Fatalf("Missing required env var: BUCKET")
	}

	commentsService := CommentsService{
		Store: CommentStore{
			Bucket:      bucket,
			Prefix:      "",
			ObjectStore: &S3ObjectStore{s3.New(session.New())},
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
			Handler: commentsService.PutComment,
		},
		pz.Route{
			Method:  "GET",
			Path:    "/posts/{post-id}/comments/{comment-id}",
			Handler: commentsService.GetComment,
		},
	))
}
