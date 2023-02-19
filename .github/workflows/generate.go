package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type PushTrigger struct {
	Branches []string `yaml:"branches,omitempty"`
	Tags     []string `yaml:"tags,omitempty"`
}

type Trigger struct {
	Push PushTrigger `yaml:"push,omitempty"`
}

type Args map[string]interface{}

type Step struct {
	Name string `yaml:"name,omitempty"`
	If   string `yaml:"if,omitempty"`
	Uses string `yaml:"uses,omitempty"`
	ID   string `yaml:"id,omitempty"`
	Run  string `yaml:"run,omitempty"`
	With Args   `yaml:"with,omitempty"`
}

type Job struct {
	RunsOn string `yaml:"runs-on"`
	Steps  []Step `yaml:"steps"`
}

type Workflow struct {
	Name string  `yaml:"name"`
	On   Trigger `yaml:"on,omitempty"`
	Jobs map[string]Job
}

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

type Image struct {
	// The name of the GitHub Action Job to build the image as well as the
	// github-username-prefixed name of the image in the registry. E.g.,
	// for github user `weberc2` and name `comments`, this would get pushed to
	// the registry as `weberc2/comments`.
	Name string

	// The path to the Dockerfile relative to the repo root.
	Dockerfile string

	// The path to the build context relative to the repo root.
	Context string

	// The build arguments.
	Args map[string]string
}

func GoImage(target string) *Image {
	return &Image{
		Name:       target,
		Context:    ".",
		Dockerfile: "./docker/golang/Dockerfile",
		Args:       map[string]string{"TARGET": target},
	}
}

func JobRelease(image *Image) Job {
	return Job{
		RunsOn: "ubuntu-latest",
		Steps: []Step{{
			Name: "Checkout",
			Uses: "actions/checkout@v2",
		}, {
			Name: "Prepare",
			ID:   "prep",
			Run: fmt.Sprintf(`DOCKER_IMAGE=${{ secrets.DOCKER_USERNAME }}/%s
VERSION=latest
SHORTREF=${GITHUB_SHA::8}

# If this is a git tag, use the tag name as a docker tag
if [[ $GITHUB_REF == refs/tags/* ]]; then
  VERSION=${GITHUB_REF#refs/tags/v}
fi
TAGS="${DOCKER_IMAGE}:${VERSION},${DOCKER_IMAGE}:${SHORTREF}"

# If the VERSION looks like a version number, assume that
# this is the most recent version of the image and also
# tag it 'latest'.
if [[ $VERSION =~ ^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$ ]]; then
  TAGS="$TAGS,${DOCKER_IMAGE}:latest"
fi

# Set output parameters.
echo ::set-output name=tags::${TAGS}
echo ::set-output name=docker_image::${DOCKER_IMAGE}`, image.Name),
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
			Uses: "docker/login-action@v1",
			With: Args{
				"username": "${{ secrets.DOCKER_USERNAME }}",
				"password": "${{ secrets.DOCKER_PASSWORD }}",
			},
		}, {
			Name: "Build",
			Uses: "docker/build-push-action@v2",
			With: Args{
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
			},
		}},
	}
}

func MarshalToWriter(w io.Writer, v interface{}) error {
	yamlEncoder := yaml.NewEncoder(w)
	yamlEncoder.SetIndent(2) // this is what you're looking for
	if err := yamlEncoder.Encode(v); err != nil {
		return fmt.Errorf("marshaling to YAML: %w", err)
	}
	return nil
}

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
			&Image{
				Name:       "linkcheck",
				Dockerfile: "docker/golang/Dockerfile",
				Context:    "mod/linkcheck",
			},
		),
	); err != nil {
		log.Fatalf("marshaling release workflow: %v", err)
	}
}
