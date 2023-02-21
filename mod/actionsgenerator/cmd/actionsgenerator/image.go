package main

import (
	"fmt"
	"log"
)

// Image represents a container image to be built.
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

// SetSinglePlatform sets the SinglePlatform field.
func (image *Image) SetSinglePlatform(platform string) *Image {
	image.SinglePlatform = platform
	return image
}

// SetDockerfile sets the Dockerfile field.
func (image *Image) SetDockerfile(dockerfile string) *Image {
	image.Dockerfile = dockerfile
	return image
}

// SetECRRegistry sets the registry to an ECR registry. The `secretPrefix`
// parameter will be prepended to `_AWS_ACCESS_KEY_ID` and
// `_AWS_SECRET_ACCESS_KEY` to determine which GitHub Action secrets to use to
// authenticate with AWS.
func (image *Image) SetECRRegistry(secretPrefix string) *Image {
	image.Registry = Registry{
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
	return image
}

// FullName gives the fully-qualified name of the image (including the registry
// prefix).
func (image *Image) FullName() string {
	if image.Registry.Type == RegistryTypeDocker {
		return fmt.Sprintf("${{ secrets.DOCKER_USERNAME }}/%s", image.Name)
	}
	if image.Registry.Type == RegistryTypeECR {
		return fmt.Sprintf("%s/%s", image.Registry.ECR.Registry, image.Name)
	}
	log.Fatalf("invalid registry type: %d", image.Registry.Type)
	return ""
}
