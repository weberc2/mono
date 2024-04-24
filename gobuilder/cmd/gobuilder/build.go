package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

type Build struct {
	Context     BuildContext     `json:"context"`
	Package     RelativePath     `json:"package"`
	Output      BuildOutput      `json:"output"`
	Environment Environment      `json:"environment"`
	Destination BuildDestination `json:"destination"`
	Stdout      io.Writer        `json:"stdout"`
	Stderr      io.Writer        `json:"stderr"`
	Stdin       io.Reader        `json:"stdin"`
}

type RelativePath string

func (path *RelativePath) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, (*string)(path)); err != nil {
		return fmt.Errorf("unmarshaling relative path: %w", err)
	}

	if strings.Contains(*(*string)(path), "..") {
		return errors.New(
			"unmarshaling relative path: illegal character sequence: `..`",
		)
	}

	return nil
}

type Environment []string

func (env *Environment) UnmarshalJSON(data []byte) error {
	var payload map[string]string
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("unmarshaling environment: %w", err)
	}
	*env = make(Environment, 0, len(payload))
	for key, value := range payload {
		*env = append(*env, fmt.Sprintf("%s=%s", key, value))
	}
	return nil
}

type BuildOutput string

func (output *BuildOutput) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, (*string)(output)); err != nil {
		return fmt.Errorf("unmarshaling build output: %w", err)
	}
	if strings.Contains(*(*string)(output), "/") {
		*output = ""
		return errors.New("unmarshaling build output: illegal character: `/`")
	}
	return nil
}
