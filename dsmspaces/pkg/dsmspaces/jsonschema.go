package dsmspaces

import "github.com/kaptinlin/jsonschema"

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

var compiler = jsonschema.NewCompiler()
