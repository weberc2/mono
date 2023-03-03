package main

// GoImage creates an `Image` parameterized for a Go module target. The target
// project should be under the root directory.
func GoImage(module, pkg string) *Image {
	return &Image{
		Name:       pkg,
		Context:    module,
		Dockerfile: "docker/golang/Dockerfile",
		Args:       map[string]string{"TARGET": pkg},
	}
}
