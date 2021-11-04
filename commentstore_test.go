package main

import (
	"testing"
)

type postStoreFake []PostID

func (psf postStoreFake) Exists(post PostID) error {
	for i := range psf {
		if psf[i] == post {
			return nil
		}
	}
	return &PostNotFoundErr{post}
}

func TestCommentStore_PutComment_ParentNotFound(t *testing.T) {
	// Given a comment store with a post called "my-post"
	commentStore := CommentStore{
		ObjectStore: objectStoreFake{},
		PostStore:   postStoreFake{"my-post"},
	}

	// When a comment is added to "my-post" with a `Parent` that doesn't exist
	_, err := commentStore.PutComment("my-post", &Comment{Parent: "doesnt-exist"})

	// Then expect a `CommentNotFoundErr` is returned
	if err, ok := err.(*CommentNotFoundErr); ok {
		if err.Post == "my-post" && err.Comment == "doesnt-exist" {
			return
		}
	}
	t.Fatalf(
		`Wanted CommentNotFoundErr{Post: "my-post", Comment: "doesnt-exist"}; found %# v`,
		err,
	)
}

func TestCommentStore_PutComment_PostNotFound(t *testing.T) {
	// Given a comment store with no posts
	commentStore := CommentStore{
		ObjectStore: objectStoreFake{},
		PostStore:   postStoreFake{},
	}

	// When a comment is added on an unknown post
	_, err := commentStore.PutComment("my-post", &Comment{})

	// Then expect a `PostNotFoundErr` is returned
	if err, ok := err.(*PostNotFoundErr); ok {
		if err.Post == "my-post" {
			return
		}
	}
	t.Fatalf(`Wanted PostNotFoundErr{Post: "my-post"}; found %# v`, err)
}

func TestCOmmentStore_ListComments_PostNotFound(t *testing.T) {
	// Given a comment store with no posts
	commentStore := CommentStore{
		ObjectStore: objectStoreFake{},
		PostStore:   postStoreFake{},
	}

	// When someone tries to list comments on a post that doesn't exist
	_, err := commentStore.PostComments("my-post", "")

	// Then expect a `PostNotFoundErr` is returned
	if err, ok := err.(*PostNotFoundErr); ok {
		if err.Post == "my-post" {
			return
		}
	}
	t.Fatalf(`Wanted PostNotFoundErr{Post: "my-post"}; found %# v`, err)
}
