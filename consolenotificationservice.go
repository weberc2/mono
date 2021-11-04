package main

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type ConsoleNotificationService struct{}

func (cns ConsoleNotificationService) Notify(u UserID, t uuid.UUID) error {
	data, err := json.Marshal(struct {
		User  UserID `json:"user"`
		Token string `json:"token"`
	}{
		User:  u,
		Token: t.String(),
	})
	if err != nil {
		return err
	}
	_, err = fmt.Printf("%s\n", data)
	return err
}
