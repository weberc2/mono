package main

import "path/filepath"

// GoImage creates an `Image` parameterized for a Go module target. The target
// project should be under the `/mod` directory.
func GoImage(target string) *Image {
	return &Image{
		Name:       target,
		Context:    filepath.Join("./mod", target),
		Dockerfile: "docker/golang/Dockerfile",
		Args:       map[string]string{"TARGET": target},
	}
}
