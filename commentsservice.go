package main

import (
	"encoding/json"

	pz "github.com/weberc2/httpeasy"
)

type CommentsService struct {
	Store CommentStore
}

func (cs *CommentsService) PutComment(r pz.Request) pz.Response {
	var c Comment
	if err := r.JSON(&c); err != nil {
		return pz.BadRequest(pz.String("Malformed `Comment` JSON"), e{err})
	}
	if user := UserID(r.Headers.Get("User")); user != c.Author {
		pz.Unauthorized(nil, struct {
			Message, Error string
		}{
			Message: "mismatch between 'User' header and 'Author' body field",
			Error:   "user %s tried posting a comment from user %s",
		})
	}
	id, err := cs.Store.PutComment(PostID(r.Vars["post-id"]), &c)
	if err != nil {
		return pz.InternalServerError(e{err})
	}
	c.ID = id
	return pz.Ok(pz.JSON(&c), struct {
		Message string
		Comment CommentID
	}{
		Message: "Created comment",
		Comment: id,
	})
}

func (cs *CommentsService) PostComments(r pz.Request) pz.Response {
	var parent CommentID
	if commentID := r.Vars["comment-id"]; commentID != "toplevel" {
		parent = CommentID(commentID)
	}
	comments, err := cs.Store.PostComments(PostID(r.Vars["post-id"]), parent)
	if err != nil {
		if err, ok := err.(*CommentNotFoundErr); ok {
			return pz.NotFound(
				pz.Stringf(
					"comment '%s' not found with post '%s'",
					r.Vars["comment-id"],
					r.Vars["post-id"],
				),
				struct {
					Message string
					Err     error
				}{
					Message: "comment not found",
					Err:     err,
				},
			)
		}
		return pz.InternalServerError(e{err})
	}
	return pz.Ok(pz.JSON(comments))
}

func (cs *CommentsService) GetComment(r pz.Request) pz.Response {
	comment, err := cs.Store.GetComment(
		PostID(r.Vars["post-id"]),
		CommentID(r.Vars["comment-id"]),
	)
	if err != nil {
		if err, ok := err.(*CommentNotFoundErr); ok {
			return pz.NotFound(
				pz.Stringf(
					"comment '%s' not found with post '%s'",
					r.Vars["comment-id"],
					r.Vars["post-id"],
				),
				struct {
					Message string
					Err     error
				}{
					Message: "comment not found",
					Err:     err,
				},
			)
		}
		return pz.InternalServerError(e{err})
	}
	return pz.Ok(pz.JSON(comment))
}

type e struct {
	err error
}

func (e e) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct{ Err string }{e.err.Error()})
}
