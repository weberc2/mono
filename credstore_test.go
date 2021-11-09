package main

import (
	"encoding/base64"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestCreate(t *testing.T) {
	const password = "oiusdpafohwerkljsfkljads;fweqr"

	var entry *UserEntry
	if err := (&CredStore{&userStoreMock{
		create: func(e *UserEntry) error { entry = e; return nil },
	}}).Create(&Credentials{
		User:     "user",
		Email:    "user@example.org",
		Password: password,
	}); err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}

	if entry == nil {
		t.Fatalf("UserStore.Create() not called or called with nil value")
	}

	if entry.User != "user" {
		t.Fatalf("UserStore.User: wanted 'user'; found '%s'", entry.User)
	}

	if entry.Email != "user@example.org" {
		t.Fatalf(
			"UserStore.Email: wanted 'user@example.org'; found '%s'",
			entry.Email,
		)
	}

	if err := bcrypt.CompareHashAndPassword(
		entry.PasswordHash,
		[]byte(password),
	); err != nil {
		t.Fatalf(
			"UserStore.PasswordHash: not generated from password '%s'",
			password,
		)
	}
}

func TestUpdate(t *testing.T) {
	const password = "oiusdpafohwerkljsfkljads;fweqr"

	var entry *UserEntry
	if err := (&CredStore{&userStoreMock{
		update: func(e *UserEntry) error { entry = e; return nil },
	}}).Update(&Credentials{
		User:     "user",
		Email:    "user@example.org",
		Password: password,
	}); err != nil {
		t.Fatalf("Unexpected err: %v", err)
	}

	if entry == nil {
		t.Fatalf("UserStore.Update() not called or called with nil value")
	}

	if entry.User != "user" {
		t.Fatalf("UserStore.User: wanted 'user'; found '%s'", entry.User)
	}

	if entry.Email != "user@example.org" {
		t.Fatalf(
			"UserStore.Email: wanted 'user@example.org'; found '%s'",
			entry.Email,
		)
	}

	if err := bcrypt.CompareHashAndPassword(
		entry.PasswordHash,
		[]byte(password),
	); err != nil {
		t.Fatalf(
			"UserStore.PasswordHash: not generated from password '%s'",
			password,
		)
	}
}

func b64(data []byte) string {
	return base64.RawStdEncoding.EncodeToString(data)
}
