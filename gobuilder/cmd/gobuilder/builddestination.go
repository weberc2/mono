package main

import "context"

type BuildDestination struct {
	Type    BuildDestinationType    `json:"type"`
	Backend BuildDestinationBackend `json:"backend"`
}

func (bd *BuildDestination) Upload(
	ctx context.Context,
	secrets *Secrets,
	directory string,
	output BuildOutput,
) error {
	return bd.Backend.Upload(ctx, secrets, directory, output)
}

type BuildDestinationBackend interface {
	Upload(
		ctx context.Context,
		secrets *Secrets,
		directory string,
		output BuildOutput,
	) error
}

type BuildDestinationType string

const (
	BuildDestinationTypeS3 BuildDestinationType = "S3"
)
