package main

import "log"

type RegistryType int

const (
	RegistryTypeDocker RegistryType = iota
	RegistryTypeECR
)

type ECRDetails struct {
	Registry string
	Username string
	Password string
}

type Registry struct {
	Type RegistryType
	ECR  ECRDetails
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
