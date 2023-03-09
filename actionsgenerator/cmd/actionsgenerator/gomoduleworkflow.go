package main

import (
	"fmt"
	"strings"
)

func GoModuleWorkflow(name string, images ...GoModuleImage) Workflow {
	out := Workflow{
		Name: name,
		On: Trigger{
			Push: PushTrigger{
				Branches: []string{"*"},
				Tags:     []string{"*"},
			},
		},
		Jobs: map[string]Job{
			"check": {
				RunsOn: "ubuntu-latest",
				Steps: []Step{
					{Uses: "actions/checkout@v3"},
					{
						Name: "Setup Go",
						Uses: "actions/setup-go@v3",
						With: Args{"go-version": "1.19.5"},
					},
					{
						Name: "Go Fmt",
						ID:   "fmt",
						Run: runInDir(
							name,
							`go install github.com/segmentio/golines@v0.10.0
output="$(golines -m 79 --shorten-comments --dry-run .)"
if [ -n "$output" ]; then
	echo "$output"
	exit 1
fi`,
						),
					},
					{
						Name: "Go Vet",
						ID:   "vet",
						Run:  runInDir(name, "go vet -v ./..."),
					},
					{
						Name: "Go Test",
						ID:   "test",
						Run:  runInDir(name, "go test -v ./..."),
					},
					{
						Name: "Go Build",
						ID:   "test",
						Run:  runInDir(name, "go build -v ./..."),
					},
				},
			},
		},
	}

	for _, img := range images {
		image := img(name)
		buildArgs := Args{
			"builder": "${{ steps.buildx.outputs.name }}",
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
			"tags":       "${{ steps.prep.outputs.tags }}",
			"cache-from": "type=gha, scope=${{ github.workflow }}",
			"cache-to":   "type=gha, scope=${{ github.workflow }}",
		}
		if image.SinglePlatform != "" {
			buildArgs["platforms"] = image.SinglePlatform
			buildArgs["provenance"] = false
		}
		out.Jobs[fmt.Sprintf("image %s", img)] = Job{
			RunsOn: "ubuntu-latest",
			Steps: []Step{{
				Name: "Checkout",
				Uses: "actions/checkout@v3",
			}, {
				Name: "Prepare",
				ID:   "prep",
				Run:  tmpl(image),
			}, {
				Name: "Set up QEMU",
				Uses: "docker/setup-qemu-action@master",
				With: Args{"platforms": "all"},
			}, {
				Name: "Set up Docker Buildx",
				ID:   "buildx",
				Uses: "docker/setup-buildx-action@master",
			}, {
				Name: "Login to DockerHub",
				If:   "github.event_name != 'pull_request'",
				Uses: "docker/login-action@v2",
				With: image.Registry.Args(),
			}, {
				Name: "Build",
				Uses: "docker/build-push-action@v4",
				With: buildArgs,
			}},
		}
	}

	return out
}

func runInDir(dir, command string) string {
	return fmt.Sprintf("(cd %s && %s)", dir, command)
}
