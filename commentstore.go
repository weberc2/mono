package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/google/uuid"
)

type PostNotFoundErr struct{ Post PostID }

func (err *PostNotFoundErr) Error() string {
	return fmt.Sprintf("post not found: %s", err.Post)
}

type PostStore interface {
	Exists(PostID) error
}

type CommentStore struct {
	ObjectStore ObjectStore
	PostStore   PostStore
	Bucket      string
	Prefix      string
}

func (cs *CommentStore) putObject(path string, data []byte) error {
	return cs.ObjectStore.PutObject(
		cs.Bucket,
		filepath.Join(cs.Prefix, path),
		bytes.NewReader(data),
	)
}

func (cs *CommentStore) getObject(key string) ([]byte, error) {
	body, err := cs.ObjectStore.GetObject(
		cs.Bucket,
		filepath.Join(cs.Prefix, key),
	)
	if err != nil {
		return nil, err
	}
	defer body.Close()
	return ioutil.ReadAll(body)
}

func (cs *CommentStore) putComment(post PostID, c *Comment) error {
	data, err := json.Marshal(&c)
	if err != nil {
		return err
	}
	if err := cs.PostStore.Exists(post); err != nil {
		return err
	}

	// If a `parent` was provided, then make sure it exists
	if c.Parent != "" {
		if _, err := cs.GetComment(post, c.Parent); err != nil {
			return err
		}
	}
	return cs.putObject(
		fmt.Sprintf("posts/%s/comments/%s/__comment__", post, c.ID),
		data,
	)
}

func (cs *CommentStore) putParentLink(post PostID, c *Comment) error {
	parent := c.Parent
	if c.Parent == "" {
		parent = "__toplevel__"
	}
	return cs.putObject(
		fmt.Sprintf("posts/%s/comments/%s/comments/%s", post, parent, c.ID),
		nil,
	)
}

func (cs *CommentStore) PutComment(post PostID, c *Comment) (CommentID, error) {
	cp := *c
	cp.ID = CommentID(uuid.NewString())
	if err := cs.putComment(post, &cp); err != nil {
		return "", err
	}
	if err := cs.putParentLink(post, &cp); err != nil {
		return "", err
	}
	return cp.ID, nil
}

func (cs *CommentStore) listObjects(prefix string) ([]string, error) {
	return cs.ObjectStore.ListObjects(
		cs.Bucket,
		filepath.Join(cs.Prefix, prefix),
	)
}

func (cs *CommentStore) getComment(key string) (Comment, error) {
	data, err := cs.getObject(key)
	if err != nil {
		return Comment{}, err
	}
	var c Comment
	err = json.Unmarshal(data, &c)
	return c, err
}

type CommentNotFoundErr struct {
	Post    PostID
	Comment CommentID
}

func (err *CommentNotFoundErr) Error() string {
	return fmt.Sprintf(
		"comment not found: post=%s comment=%s",
		err.Post,
		err.Comment,
	)
}

func (cs *CommentStore) GetComment(post PostID, comment CommentID) (Comment, error) {
	key := fmt.Sprintf("posts/%s/comments/%s/__comment__", post, comment)
	c, err := cs.getComment(key)
	if err != nil {
		if _, ok := err.(*ObjectNotFoundErr); ok {
			return Comment{}, &CommentNotFoundErr{Post: post, Comment: comment}
		}
		return Comment{}, err
	}
	return c, nil
}

func (cs *CommentStore) PostComments(post PostID, parent CommentID) ([]Comment, error) {
	if parent == "" {
		parent = "__toplevel__"
	}

	prefix := fmt.Sprintf("posts/%s/comments/%s/comments/", post, parent)
	keys, err := cs.listObjects(prefix)
	if err != nil {
		return nil, fmt.Errorf(
			"listing objects with prefix '%s': %w",
			prefix,
			err,
		)
	}

	comments := make([]Comment, len(keys))
	for i, key := range keys {
		comment, err := cs.GetComment(post, CommentID(filepath.Base(key)))
		if err != nil {
			return nil, err
		}
		comments[i] = comment
	}

	return comments, nil
}
