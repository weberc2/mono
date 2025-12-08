package main

type Registry int

const (
	RegistryDocker Registry = iota
	RegistryGHCR
)

var (
	RegistryTitles = [...]string{
		RegistryDocker: "DockerHub",
		RegistryGHCR:   "GitHub Container Registry",
	}

	RegistryPrefixes = [...]string{
		RegistryDocker: "${{ secrets.DOCKER_USERNAME }}",
		RegistryGHCR:   "ghcr.io/${{ github.actor }}",
	}

	RegistryArgs = [...]Args{
		RegistryDocker: {
			"username": "${{ secrets.DOCKER_USERNAME }}",
			"password": "${{ secrets.DOCKER_PASSWORD }}",
		},
		RegistryGHCR: {
			"registry": "ghcr.io",
			"username": "${{ github.actor }}",
			"password": "${{ secrets.GITHUB_TOKEN }}",
		},
	}
)
