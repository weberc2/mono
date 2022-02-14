package main

import (
	"fmt"
	"io"
	"log"
	"os"

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

func WorkflowRelease(targets ...string) Workflow {
	jobs := make(map[string]Job, len(targets))
	for _, target := range targets {
		jobs[target] = JobRelease(target)
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

func JobRelease(target string) Job {
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
echo ::set-output name=docker_image::${DOCKER_IMAGE}`, target),
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
				"builder":    "${{ steps.buildx.outputs.name }}",
				"build-args": "TARGET=" + target,
				"context":    ".",
				"file":       "./Dockerfile",
				"platforms":  "linux/amd64,linux/arm64",
				"push":       true,
				"tags":       "${{ steps.prep.outputs.tags }}",
				"cache-from": "type=gha, scope=${{ github.workflow }}",
				"cache-to":   "type=gha, scope=${{ github.workflow }}",
			},
		}},
	}
}

func MarshalTo(dst string, v interface{}) error {
	f, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("opening file `%s`: %w", dst, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Println("WARNING failed to close file `%s`: %v", dst, err)
		}
	}()
	return MarshalToWriter(f, v)
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
		WorkflowRelease("auth", "comments"),
	); err != nil {
		log.Fatalf("marshaling release workflow: %v", err)
	}
}
