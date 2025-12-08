package main

import (
	"log"
	"os"
)

func main() {
	if err := MarshalToWriter(
		os.Stdout,
		WorkflowRelease(
			GoImage("serverstatus", "serverstatus").SetRegistry(RegistryGHCR),
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
