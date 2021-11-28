package types

import (
	pz "github.com/weberc2/httpeasy"
)

type UserID string

type UserEntry struct {
	User         UserID
	Email        string
	PasswordHash []byte
}

type UserStore interface {
	Get(UserID) (*UserEntry, error)
	Create(*UserEntry) error
	Upsert(*UserEntry) error
}

var ErrUserNotFound = &pz.HTTPError{Status: 404, Message: "User not found"}
