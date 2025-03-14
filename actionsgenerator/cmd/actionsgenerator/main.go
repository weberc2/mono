package main

import (
	"log"
	"os"
)

func main() {
	if err := MarshalToWriter(
		os.Stdout,
		WorkflowRelease(
			&Image{
				Name: "rain",
				Dockerfile: "./docker/rain/Dockerfile",
				Context: "./docker/rain",
			},
			&Image{
				Name:       "pgbackup",
				Dockerfile: "./docker/pgbackup/Dockerfile",
				Context:    "./docker/pgbackup",
			},
			GoImage("analytics", "analytics").
				SetRegistry(&Registry{
					Type:           RegistryTypeECR,
					ID:             "988080168334.dkr.ecr.us-east-2.amazonaws.com",
					UsernameSecret: "ANALYTICS_AWS_ACCESS_KEY_ID",
					PasswordSecret: "ANALYTICS_AWS_SECRET_ACCESS_KEY",
				}).
				SetSinglePlatform("linux/arm64"),
			GoImage("comments", "auth"),
			GoImage("comments", "tokens"),
			GoImage("comments", "users"),
			GoImage("comments", "comments"),
			GoImage("linkcheck", "linkcheck"),
			GoImage("gobuilder", "gobuilder").
				// Use the Dockerfile in the module directory rather than the
				// default Go Dockerfile (the gobuilder Dockerfile preserves
				// the Go toolchain in the final image so it can build other
				// images).
				SetDockerfile("docker/gobuilder/Dockerfile").
				SetRegistry(&Registry{
					Type:           RegistryTypeECR,
					ID:             "988080168334.dkr.ecr.us-east-2.amazonaws.com",
					UsernameSecret: "GOBUILDER_AWS_ACCESS_KEY_ID",
					PasswordSecret: "GOBUILDER_AWS_SECRET_ACCESS_KEY",
				}).
				// disable multiarch for lambda (lambda can't run multiarch
				// containers yet).
				SetSinglePlatform("linux/amd64"),
		),
	); err != nil {
		log.Fatalf("marshaling release workflow: %v", err)
	}
}
