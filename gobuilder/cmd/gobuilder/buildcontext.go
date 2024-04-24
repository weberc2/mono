package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

type BuildContext struct {
	Type    BuildContextType    `json:"type"`
	Backend BuildContextBackend `json:"backend"`
}

func (bc *BuildContext) UnmarshalJSON(data []byte) error {
	var payload struct {
		Type    BuildContextType `json:"type"`
		Backend json.RawMessage  `json:"backend"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("unmarshaling build context: %w", err)
	}
	switch payload.Type {
	case BuildContextTypeGit:
		var git BuildContextGit
		if err := json.Unmarshal(payload.Backend, &git); err != nil {
			return fmt.Errorf("unmarshaling git build context: %w", err)
		}
		bc.Backend = &git
		return nil
	default:
		return errors.New(
			"unmarshaling build context: unsupported build context type",
		)
	}
}

func (bc *BuildContext) CopyTo(
	ctx context.Context,
	secrets *Secrets,
	directory string,
) error {
	return bc.Backend.CopyToBuildDirectory(ctx, secrets, directory)
}

type BuildContextBackend interface {
	CopyToBuildDirectory(
		ctx context.Context,
		secrets *Secrets,
		directory string,
	) error
}

type BuildContextType string

const (
	BuildContextTypeReader BuildContextType = "READER"
	BuildContextTypeGit    BuildContextType = "GIT"
)
