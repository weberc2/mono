package comments

import (
	"fmt"
	"time"

	"html"

	"github.com/weberc2/comments/pkg/types"
	pz "github.com/weberc2/httpeasy"
)

const (
	bodySizeMin = 8
	bodySizeMax = 2056
)

var (
	ErrInvalidPost  = &pz.HTTPError{Status: 400, Message: "invalid post"}
	ErrBodyTooShort = &pz.HTTPError{Status: 400, Message: "body too short"}
	ErrBodyTooLong  = &pz.HTTPError{Status: 400, Message: "body too long"}
)

type CommentsModel struct {
	types.CommentsStore
	IDFunc   func() types.CommentID
	TimeFunc func() time.Time
}

func (cm *CommentsModel) Put(c *types.Comment) (*types.Comment, error) {
	if c.Post == "" {
		return nil, ErrInvalidPost
	}
	if len(c.Body) < bodySizeMin {
		return nil, ErrBodyTooShort
	}
	if len(c.Body) > bodySizeMax {
		return nil, ErrBodyTooLong
	}
	if c.Parent != "" {
		parent, err := cm.CommentsStore.Comment(
			c.Post,
			c.Parent,
		)
		if err != nil {
			return nil, fmt.Errorf("fetching parent comment: %w", err)
		}
		if parent.Deleted {
			return nil, fmt.Errorf(
				"fetching parent comment: %w",
				&types.CommentNotFoundErr{Post: c.Post, Comment: c.Parent},
			)
		}
	}
	now := cm.TimeFunc()
	cp := *c
	cp.ID = cm.IDFunc()
	cp.Created = now
	cp.Modified = now
	cp.Deleted = false
	cp.Body = html.EscapeString(c.Body)
	return cm.CommentsStore.Put(&cp)
}

func (cm *CommentsModel) Delete(p types.PostID, c types.CommentID) error {
	if err := cm.CommentsStore.Update(
		types.NewCommentPatch(c, p).SetDeleted(true).
			SetModified(cm.TimeFunc()),
	); err != nil {
		return fmt.Errorf("soft-deleting comment: %w", err)
	}
	return nil
}

func (cm *CommentsModel) Replies(
	post types.PostID,
	parent types.CommentID,
) ([]*types.Comment, error) {
	comments, err := cm.CommentsStore.Replies(post, parent)
	if err != nil {
		return nil, err
	}

	for _, comment := range comments {
		if comment.Deleted {
			// Redact author and body fields from deleted comments
			comment.Author = ""
			comment.Body = ""
		}
	}

	return comments, nil
}
