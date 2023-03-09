package main

import (
	"log"
	"os"
)

var workflows = []Workflow{
	GoModuleWorkflow(
		"comments",
		GoModImage("auth"),
		GoModImage("tokens"),
		GoModImage("users"),
		GoModImage("comments"),
	),
	GoModuleWorkflow("linkcheck", GoModImage("linkcheck")),
	GoModuleWorkflow(
		"gobuilder",
		GoModImage("gobuilder").
			// Use the Dockerfile in the module directory rather than the
			// default Go Dockerfile (the gobuilder Dockerfile preserves
			// the Go toolchain in the final image so it can build other
			// images).
			SetDockerfile("gobuilder/Dockerfile").
			SetRegistry(&Registry{
				Type:           RegistryTypeECR,
				ID:             "988080168334.dkr.ecr.us-east-2.amazonaws.com",
				UsernameSecret: "GOBUILDER_AWS_ACCESS_KEY_ID",
				PasswordSecret: "GOBUILDER_AWS_SECRET_ACCESS_KEY",
			}).
			// disable multiarch for lambda
			SetSinglePlatform("linux/amd64"),
	),
	ImageWorkflow(&Image{
		Name:       "pgbackup",
		Dockerfile: "./docker/pgbackup/Dockerfile",
		Context:    "./docker/pgbackup",
	}),
	ImageWorkflow(&Image{
		Name:       "blog",
		Dockerfile: "./docker/blog/Dockerfile",
		Context:    "blog",
	}),
}

func ImageWorkflow(image *Image) Workflow {
	return *NewWorkflow("pgbackup").
		WithJob("build/push image", JobRelease(image))
}

func main() {
	for _, workflow := range workflows {
		if err := MarshalWorkflow(os.Args[1], &workflow); err != nil {
			log.Fatal(err)
		}
	}
}
