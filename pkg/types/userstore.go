package types

import (
	pz "github.com/weberc2/httpeasy"
)

type UserID string

type UserEntry struct {
	User         UserID `json:"user"`
	Email        string `json:"email"`
	PasswordHash []byte `json:"-"`
}

type UserStore interface {
	Get(UserID) (*UserEntry, error)
	Create(*UserEntry) error
	Upsert(*UserEntry) error
}

var ErrUserNotFound = &pz.HTTPError{Status: 404, Message: "User not found"}
