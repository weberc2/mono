package main

import (
	"fmt"
	"strings"
	"text/template"
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
	return Job{
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

var releaseJobPrepareStepScriptTemplate = template.Must(
	template.
		New("").
		// use different delims so GH's ${{ secrets.XYZ }} syntax doesn't
		// collide
		Delims("{%", "%}").
		Parse(
			`DOCKER_IMAGE={% .DockerImage %}
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
echo ::set-output name=docker_image::${DOCKER_IMAGE}`))

func tmpl(image *Image) string {
	var sb strings.Builder
	if err := releaseJobPrepareStepScriptTemplate.Execute(
		&sb,
		image,
	); err != nil {
		panic(err)
	}
	return sb.String()
}
