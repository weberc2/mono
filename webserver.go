package main

import (
	"errors"
	html "html/template"

	pz "github.com/weberc2/httpeasy"
)

type logging struct {
	Post   PostID    `json:"post"`
	Parent CommentID `json:"parent"`
	User   UserID    `json:"user,omitempty"`
	Error  string    `json:"error,omitempty"`
}

type WebServer struct {
	LoginURL                   string
	LogoutURL                  string
	BaseURL                    string
	Comments                   CommentStore
	RepliesTemplate            *html.Template
	DeleteConfirmationTemplate *html.Template
}

func (ws *WebServer) Replies(r pz.Request) pz.Response {
	post := PostID(r.Vars["post-id"])
	parent := CommentID(r.Vars["parent-id"])
	user := UserID(r.Headers.Get("User"))
	if parent == "toplevel" {
		parent = "" // this tells the CommentStore to fetch toplevel replies.
	}
	replies, err := ws.Comments.Replies(post, parent)
	if err != nil {
		var c *CommentNotFoundErr
		if errors.As(err, &c) {
			pz.NotFound(nil, &logging{
				Post:   post,
				Parent: parent,
				User:   user,
				Error:  err.Error(),
			})
		}

		return pz.InternalServerError(&logging{
			Post:   post,
			Parent: parent,
			User:   user,
			Error:  err.Error(),
		})
	}

	return pz.Ok(
		pz.HTMLTemplate(ws.RepliesTemplate, struct {
			LoginURL  string    `json:"loginURL"`
			LogoutURL string    `json:"logoutURL"`
			BaseURL   string    `json:"baseURL"`
			Post      PostID    `json:"post"`
			Parent    CommentID `json:"parent"`
			Replies   []Comment `json:"replies"`
			User      UserID    `json:"user"`
		}{
			LoginURL:  ws.LoginURL,
			LogoutURL: ws.LogoutURL,
			BaseURL:   ws.BaseURL,
			Post:      post,
			Parent:    parent,
			Replies:   replies,
			User:      user, // empty if unauthorized
		}),
		&logging{Post: post, Parent: parent, User: user},
	)
}

func (ws *WebServer) DeleteConfirm(r pz.Request) pz.Response {
	params := struct {
		BaseURL string  `json:"baseURL"`
		User    UserID  `json:"user"`
		Post    PostID  `json:"post"`
		Comment Comment `json:"comment"`
		Error   string  `json:"error,omitempty"`
	}{
		BaseURL: ws.BaseURL,
		Post:    PostID(r.Vars["post-id"]),
		Comment: Comment{ID: CommentID(r.Vars["comment-id"])},
		User:    UserID(r.Headers.Get("User")), // empty if unauthorized
	}

	comment, err := ws.Comments.Comment(params.Post, params.Comment.ID)
	if err != nil {
		var e *CommentNotFoundErr
		if errors.As(err, &e) {
			params.Error = err.Error()
			return pz.NotFound(nil, params)
		}
	}

	params.Comment = comment

	return pz.Ok(
		pz.HTMLTemplate(ws.DeleteConfirmationTemplate, params),
		params,
	)
}
