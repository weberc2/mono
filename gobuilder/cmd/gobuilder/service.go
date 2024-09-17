package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Service struct {
	Secrets              *Secrets `json:"-"`
	BuildParentDirectory string   `json:"buildParentDirectory"`
}

func (s *Service) Build(ctx context.Context, build *Build) error {
	directory, err := s.prepareBuildDirectory(ctx, build)
	if err != nil {
		return fmt.Errorf("building `%s`: %w", build.Output, err)
	}

	cmd := exec.Command(
		"go",
		"build",
		"-o", string(build.Output),
		string(build.Package),
	)
	cmd.Dir = directory
	cmd.Env = append(os.Environ(), []string(build.Environment)...)
	cmd.Stderr = build.Stderr
	cmd.Stdout = build.Stdout
	cmd.Stdin = build.Stdin

	if err = cmd.Run(); err != nil {
		err = fmt.Errorf("running `go build` command: %w", err)
		goto CLEANUP
	}

	if err = build.Destination.Upload(
		ctx,
		s.Secrets,
		directory,
		build.Output,
	); err != nil {
		goto CLEANUP
	}

CLEANUP:
	// if e := os.RemoveAll(directory); e != nil {
	// 	err = fmt.Errorf("cleaning up build directory: %w", e)
	// }

	if err != nil {
		err = fmt.Errorf("building `%s`: %w", build.Output, err)
	}

	return err
}

func (s *Service) prepareBuildDirectory(
	ctx context.Context,
	build *Build,
) (string, error) {
	directory := s.buildDirectory(build.Output)

	if err := build.Context.CopyTo(ctx, s.Secrets, directory); err != nil {
		return "", fmt.Errorf("preparing build directory: %w", err)
	}

	return directory, nil
}

func (s *Service) buildDirectory(output BuildOutput) string {
	var data [8]byte
	rand.Read(data[:])
	return filepath.Join(
		s.BuildParentDirectory,
		fmt.Sprintf(
			"%s-%s",
			output,
			base64.RawStdEncoding.EncodeToString(data[:]),
		),
	)
}
