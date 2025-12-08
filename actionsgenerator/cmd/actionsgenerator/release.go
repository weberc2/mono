package main

import (
	"fmt"
	"strings"
)

// WorkflowRelease builds a job for releasing container images.
func WorkflowRelease(images ...*Image) Workflow {
	jobs := make(map[string]Job, len(images))
	for _, image := range images {
		jobs[image.Name] = JobRelease(image)
	}
	return Workflow{
		Name: "release",
		On: Trigger{
			Push: PushTrigger{
				Branches: []string{"*"},
				Tags:     []string{"*"},
			},
		},
		Jobs: jobs,
	}
}

// JobRelease builds a job for releasing container images.
func JobRelease(image *Image) Job {
	buildArgs := Args{
		"build-args": func() string {
			lines := make([]string, 0, len(image.Args))
			for key, value := range image.Args {
				lines = append(lines, fmt.Sprintf("%s=%s", key, value))
			}
			return strings.Join(lines, "\n")
		}(),
		"context":   image.Context,
		"file":      image.Dockerfile,
		"platforms": "linux/amd64,linux/arm64",
		"push":      true,
		"tags": fmt.Sprintf(
			"%[1]s:${{ github.sha }}\n%[1]s:latest",
			image.FullName(),
		),
		"cache-to":   "type=gha,mode=max,scope=go-cache",
		"cache-from": "type=gha,scope=go-cache",
	}
	if image.SinglePlatform != "" {
		buildArgs["platforms"] = image.SinglePlatform
		buildArgs["provenance"] = false
	}
	return Job{
		RunsOn: "ubuntu-latest",
		Steps: []Step{{
			Uses: "actions/checkout@v4",
		}, {
			Name: "Set up QEMU",
			Uses: "docker/setup-qemu-action@v3",
		}, {
			Name: "Set up Docker Buildx",
			Uses: "docker/setup-buildx-action@v3",
		}, {
			Name: fmt.Sprintf("Login to %s", RegistryTitles[image.Registry]),
			If:   "github.event_name != 'pull_request'",
			Uses: "docker/login-action@v3",
			With: RegistryArgs[image.Registry],
		}, {
			Name: "Build and push",
			Uses: "docker/build-push-action@v5",
			With: buildArgs,
		}},
	}
}
