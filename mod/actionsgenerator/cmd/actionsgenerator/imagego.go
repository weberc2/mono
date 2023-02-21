package main

import "path/filepath"

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
