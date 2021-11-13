package main

import (
	"errors"
	"fmt"

	"github.com/nbutton23/zxcvbn-go"

	"golang.org/x/crypto/bcrypt"
)

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

var ErrPasswordTooSimple = errors.New("Password is too simple")

type CredStore struct {
	Users UserStore
}

func (cs *CredStore) Validate(creds *Credentials) error {
	entry, err := cs.Users.Get(creds.User)
	if err != nil {
		// If the user doesn't exist, we want to return ErrCredentials in order
		// to minimize the information we give to potential attackers.
		if errors.Is(err, ErrUserNotFound) {
			return ErrCredentials
		}
		return fmt.Errorf("validating credentials: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword(
		entry.PasswordHash,
		[]byte(creds.Password),
	); err != nil {
		return ErrCredentials
	}
	return nil
}

func validatePassword(creds *Credentials) error {
	minEntropyMatch := zxcvbn.PasswordStrength(
		creds.Password,
		[]string{string(creds.User), creds.Email},
	)
	if minEntropyMatch.Score < 3 {
		return fmt.Errorf("validating password: %w", ErrPasswordTooSimple)
	}
	return nil
}

func makeUserEntry(creds *Credentials) (*UserEntry, error) {
	if err := validatePassword(creds); err != nil {
		return nil, err
	}
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(creds.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, err
	}
	return &UserEntry{
		User:         creds.User,
		Email:        creds.Email,
		PasswordHash: hashedPassword,
	}, nil
}

func (cs *CredStore) Create(creds *Credentials) error {
	entry, err := makeUserEntry(creds)
	if err != nil {
		return fmt.Errorf("creating credentials: %w", err)
	}

	if err := cs.Users.Create(entry); err != nil {
		return fmt.Errorf("creating credentials: %w", err)
	}
	return nil
}

func (cs *CredStore) Upsert(creds *Credentials) error {
	entry, err := makeUserEntry(creds)
	if err != nil {
		return fmt.Errorf("creating user entry: %w", err)
	}

	if err := cs.Users.Upsert(entry); err != nil {
		return fmt.Errorf("upserting user store: %w", err)
	}

	return nil
}
