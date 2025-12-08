package main

import (
	"log"
	"os"
)

func main() {
	if err := MarshalToWriter(
		os.Stdout,
		WorkflowRelease(
			GoImage("kubestatus", "kubestatus").
				SetRegistry(RegistryGHCR).
				SetSinglePlatform("linux/amd64"), // TODO: remove this--for testing only
			&Image{
				Name:       "rain",
				Dockerfile: "./docker/rain/Dockerfile",
				Context:    "./docker/rain",
			},
			GoImage("linkcheck", "linkcheck"),
		),
	); err != nil {
		log.Fatalf("marshaling release workflow: %v", err)
	}
}
