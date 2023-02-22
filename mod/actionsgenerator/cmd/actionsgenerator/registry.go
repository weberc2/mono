package main

import "fmt"

// RegistryType is the type of ImageRegistry
type RegistryType int

const (
	// RegistryTypeDocker indicates that the registry is Docker Inc's hub.
	RegistryTypeDocker RegistryType = iota

	// RegistryTypeECR indicates that the registry is AWS's Elastic Container
	// Registry (ECR).
	RegistryTypeECR
)

// ECRDetails holds the identifier and credentials for the ECR registry.
type ECRDetails struct {
	// Registry is the identifier for the ECR registry.
	Registry string

	// Username is the username for the ECR registry. This should always be a
	// string like `"${{ secrets.APP_NAME_AWS_ACCESS_KEY_ID}}`, which tells
	// GitHub Actions to pull the actual username value out of its secrets
	// store.
	Username string

	// Password is the password for the ECR registry. This should always be a
	// string like `"${{ secrets.APP_NAME_AWS_SECRET_ACCESS_KEY }}`, which
	// tells GitHub Actions to pull the actual password value out of its
	// secrets store.
	Password string
}

// Registry holds the details for pushing to the target registry.
type Registry struct {
	// Type indicates the type of the registry (ECR, Docker, etc).
	Type RegistryType

	// ID holds the identifier for the target ECR registry. This is optional
	// only for the official Docker registry (not for private instances of the
	// Docker registry).
	ID string

	// UsernameSecret holds the name of the secret to be provided as the
	// `username` field to the docker/build-push-image action. For ECR registry
	// types, this should hold the name of a secret containing the access key
	// ID for the AWS IAM user with permissions to push the image. If this is
	// omitted, a default secret name will be assumed based on the registry
	// type:
	//
	// * RegistryTypeDocker: DOCKER_USERNAME
	// * RegistryTypeECR: AWS_ACCESS_KEY_ID
	UsernameSecret string

	// PasswordSecret holds the name of the secret to be provided as the
	// `password` field to the docker/build-push-image action. For ECR registry
	// types, this should hold the name of a secret containing the secret
	// access key for the AWS IAM user with permissions to push the image. If
	// this is omitted, a default secret name will be assumed based on the
	// registry type:
	//
	// * RegistryTypeDocker: DOCKER_PASSWORD
	// * RegistryTypeECR: AWS_SECRET_ACCESS_KEY
	PasswordSecret string
}

// Args builds an `Args` which will be dropped into the docker/build-push-image
// action configuration.
func (r *Registry) Args() Args {
	username := r.UsernameSecret
	password := r.PasswordSecret
	if username == "" {
		username = defaultUsernameSecrets[r.Type]
	}
	if password == "" {
		password = defaultPasswordSecrets[r.Type]
	}
	args := Args{
		"username": fmt.Sprintf("${{ secrets.%s }}", username),
		"password": fmt.Sprintf("${{ secrets.%s }}", password),
	}
	if r.Type == RegistryTypeDocker && r.ID == "" {
		return args
	}
	args["registry"] = r.ID
	return args
}

var defaultUsernameSecrets = [...]string{
	RegistryTypeDocker: "DOCKER_USERNAME",
	RegistryTypeECR:    "AWS_ACCESS_KEY_ID",
}

var defaultPasswordSecrets = [...]string{
	RegistryTypeDocker: "DOCKER_PASSWORD",
	RegistryTypeECR:    "AWS_SECRET_ACCESS_KEY",
}
