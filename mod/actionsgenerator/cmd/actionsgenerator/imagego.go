package main

import "path/filepath"

// GoImage creates an `Image` parameterized for a Go module target. The target
// project should be under the `/mod` directory.
func GoImage(module, pkg string) *Image {
	return &Image{
		Name:       pkg,
		Context:    filepath.Join("./mod", module),
		Dockerfile: "docker/golang/Dockerfile",
		Args:       map[string]string{"TARGET": pkg},
	}
}
