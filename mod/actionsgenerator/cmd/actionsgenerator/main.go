package main

import (
	"log"
	"os"
)

func main() {
	if err := MarshalToWriter(
		os.Stdout,
		WorkflowRelease(
			GoImage("auth"),
			GoImage("comments"),
			&Image{
				Name:       "pgbackup",
				Dockerfile: "./docker/pgbackup/Dockerfile",
				Context:    "./docker/pgbackup",
			},
			GoModImage("linkcheck"),
			GoModImage("gobuilder").
				// Use the Dockerfile in the module directory rather than the
				// default Go Dockerfile (the gobuilder Dockerfile preserves
				// the Go toolchain in the final image so it can build other
				// images).
				SetDockerfile("mod/gobuilder/Dockerfile").
				SetECRRegistry("GOBUILDER").
				// disable multiarch for lambda
				SetSinglePlatform("linux/amd64"),
		),
	); err != nil {
		log.Fatalf("marshaling release workflow: %v", err)
	}
}
