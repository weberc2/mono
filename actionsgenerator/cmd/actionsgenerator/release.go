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
		"context":    image.Context,
		"file":       image.Dockerfile,
		"platforms":  "linux/amd64,linux/arm64",
		"push":       true,
		"cache-from": fmt.Sprintf("type=registry,ref=%s:cache", image.FullName()),
		"cache-to":   fmt.Sprintf("type=registry,ref=%s:cache,mode=max", image.FullName()),
		"tags": fmt.Sprintf(
			"%[1]s:${{ github.sha }}\n%[1]s:latest",
			image.FullName(),
		),
	}
	if image.SinglePlatform != "" {
		buildArgs["platforms"] = image.SinglePlatform
		buildArgs["provenance"] = false
	}
	return Job{
		RunsOn: "ubuntu-latest",
		Steps: []Step{{
			Uses: "actions/checkout@v4",
			// }, {
			// 	Name: "Setup QEMU",
			// 	Uses: "docker/setup-qemu-action@v3",
		}, {
			Name: "Setup Docker Buildx",
			Uses: "docker/setup-buildx-action@v3",
		}, {
			Name: "Setup Go Build Cache",
			Uses: "actions/cache@v4",
			With: Args{
				"path": `
/tmp/go-cache/cache
/tmp/go-cache/mod`,
				"key": "go-cache", // fixed key, always use the same cache
			},
		}, {
			Name: "Debug",
			Run:  "mkdir -p /tmp/go-cache && ls -l /tmp/go-cache",
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
