package main

import "log"

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

	// ECRDetails holds additional parameters for authenticating with an ECR
	// registry. It's empty unless the type is `RegistryTypeECR`.
	ECR ECRDetails
}

// Args builds an `Args` which will be dropped into the docker/build-push-image
// action configuration.
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
