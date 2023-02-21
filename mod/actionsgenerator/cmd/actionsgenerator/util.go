package main

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

func MarshalToWriter(w io.Writer, v interface{}) error {
	yamlEncoder := yaml.NewEncoder(w)
	yamlEncoder.SetIndent(2) // this is what you're looking for
	if err := yamlEncoder.Encode(v); err != nil {
		return fmt.Errorf("marshaling to YAML: %w", err)
	}
	return nil
}
