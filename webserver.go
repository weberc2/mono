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
	LoginURL  string
	LogoutURL string
	BaseURL   string
	Comments  CommentStore
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
		pz.HTMLTemplate(repliesTemplate, struct {
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
	return pz.Ok(pz.HTMLTemplate(deleteConfirmationTemplate, params), params)
}

var (
	repliesTemplate = html.Must(html.New("").Parse(`<html>
<head></head>
<body>
<h1>Replies</h1>
<div id=replies>
{{if .User}}
    {{.User}} - <a href="{{.LogoutURL}}">logout</a>
{{else}}
    <a href="{{.LoginURL}}">login</a>
{{end}}

{{$baseURL := .BaseURL}}
{{$post := .Post}}
{{$user := .User}}
{{range .Replies}}
	<div id="{{.ID}}">
		<div class="comment-header">
			<span class="author">{{.Author}}</p>
			<span class="date">{{.Created}}</p>
			{{if eq .Author $user}}
			<a href="{{$baseURL}}/posts/{{$post}}/comments/{{.ID}}/delete-confirm">
				delete
			</a>
			<a href="{{$baseURL}}/posts/{{$post}}/comments/{{.ID}}/edit">
				edit
			</a>
			{{end}}
			{{/* if the user is logged in they can reply */}}
			{{if $user}}
			<a href="{{$baseURL}}/posts/{{$post}}/comments/{{.ID}}/reply">
				reply
			</a>
			{{end}}
			<p class="body">{{.Body}}</p>
		</div>
	</div>
{{end}}
</div>
</body>
</html>`))

	deleteConfirmationTemplate = html.Must(html.New("").Parse(`<html>
<head></head>
<body>
<h1>Confirm Comment Deletion</h1>
<div id="comment">
    {{.Comment.Body}}
</div>
<div id="cancel">
    {{/*
       * For now, return to the comment itself. In the future we may pass a
       * return location through in case we have multiple delete comment
       * buttons.
    */}}
    <a href="{{.BaseURL}}/posts/{{.Post}}/comments/{{.Comment.ID}}">Cancel</a>
</div>
<div id="delete">
    <a href="{{.BaseURL}}/posts/{{.Post}}/comments/{{.Comment.ID}}/delete">
        Delete
    </a>
</div>
</div>
</body>
</html>`))
)
