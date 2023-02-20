package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
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

	// SinglePlatform is an optional field. Setting it to ${os}/${arch} will
	// disable multiarch support. This is used for building AWS Lambda Function
	// containers, since AWS Lambda does not support multiarch images at this
	// time. Example: `linux/amd64`
	SinglePlatform string

	Registry Registry
}

func (i *Image) SetSinglePlatform(platform string) *Image {
	i.SinglePlatform = platform
	return i
}

func (i *Image) SetECRRegistry(secretPrefix string) *Image {
	i.Registry = Registry{
		Type: RegistryTypeECR,
		ECR: ECRDetails{
			Registry: "988080168334.dkr.ecr.us-east-2.amazonaws.com",
			Username: fmt.Sprintf(
				"${{ secrets.%s_AWS_ACCESS_KEY_ID }}",
				secretPrefix,
			),
			Password: fmt.Sprintf(
				"${{ secrets.%s_AWS_SECRET_ACCESS_KEY }}",
				secretPrefix,
			),
		},
	}
	return i
}

type RegistryType int

const (
	RegistryTypeDocker RegistryType = iota
	RegistryTypeECR
)

type Registry struct {
	Type RegistryType
	ECR  ECRDetails
}

type ECRDetails struct {
	Registry string
	Username string
	Password string
}

func (r *Registry) Args() Args {
	if r.Type == RegistryTypeDocker {
		return Args{
			"username": "${{ secrets.DOCKER_USERNAME }}",
			"password": "${{ secrets.DOCKER_PASSWORD }}",
		}
	}
	if r.Type == RegistryTypeECR {
		return Args{
			"registry": r.ECR.Registry,
			"username": r.ECR.Username,
			"password": r.ECR.Password,
		}
	}
	log.Fatalf("invalid registry type: %d", r.Type)
	return nil
}

func (image *Image) DockerImage() string {
	if image.Registry.Type == RegistryTypeDocker {
		return fmt.Sprintf("${{ secrets.DOCKER_USERNAME }}/%s", image.Name)
	}
	if image.Registry.Type == RegistryTypeECR {
		return fmt.Sprintf("%s/%s", image.Registry.ECR.Registry, image.Name)
	}
	log.Fatalf("invalid registry type: %d", image.Registry.Type)
	return ""
}

func GoImage(target string) *Image {
	return &Image{
		Name:       target,
		Context:    ".",
		Dockerfile: "./docker/golang/Dockerfile",
		Args:       map[string]string{"TARGET": target},
	}
}

func GoModImage(target string) *Image {
	return &Image{
		Name:       target,
		Context:    filepath.Join("./mod", target),
		Dockerfile: "docker/golang/Dockerfile",
		Args:       map[string]string{"TARGET": target},
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
			Uses: "actions/checkout@v2",
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
			GoModImage("linkcheck"),
			GoModImage("gobuilder").
				SetECRRegistry("GOBUILDER").
				// disable multiarch for lambda
				SetSinglePlatform("linux/amd64"),
		),
	); err != nil {
		log.Fatalf("marshaling release workflow: %v", err)
	}
}
