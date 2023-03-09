package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func MarshalWorkflow(workflowDir string, workflow *Workflow) error {
	filePath := filepath.Join(workflowDir, workflow.Name+".yaml")
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf(
			"marshaling workflow `%s`: opening output file `%s`: %w",
			workflow.Name,
			filePath,
			err,
		)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf(
				"ERROR closing output file for workflow `%s`: %v",
				workflow.Name,
				err,
			)
		}
	}()

	if err := MarshalToWriter(file, workflow); err != nil {
		return fmt.Errorf("marshaling workflow `%s`: %w", workflow.Name, err)
	}

	return nil
}

// MarshalToWriter marshals `v` and writes the result directly to `w`.
func MarshalToWriter(w io.Writer, v interface{}) error {
	yamlEncoder := yaml.NewEncoder(w)
	yamlEncoder.SetIndent(2) // this is what you're looking for
	if err := yamlEncoder.Encode(v); err != nil {
		return fmt.Errorf("marshaling to YAML: %w", err)
	}
	return nil
}
