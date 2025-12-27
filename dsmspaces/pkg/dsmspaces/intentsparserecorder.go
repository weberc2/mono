package dsmspaces

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/openai/openai-go/packages/param"
)

type IntentsParseRecorder interface {
	RecordIntentsParse(
		ctx context.Context,
		query string,
		intents json.RawMessage,
		prompt string, // TODO: in the future replace this with a prompt ID
		model string,
		temperature param.Opt[float64],
	) (err error)
}

type NullIntentsParseRecorder struct{}

var _ IntentsParseRecorder = NullIntentsParseRecorder{}

func (NullIntentsParseRecorder) RecordIntentsParse(
	ctx context.Context,
	query string,
	intents json.RawMessage,
	prompt string,
	model string,
	temperature param.Opt[float64],
) (err error) {
	return
}

type FileIntentsParseRecorder struct {
	lock sync.Mutex
	path string
}

func NewFileIntentsParseRecorder(path string) *FileIntentsParseRecorder {
	return &FileIntentsParseRecorder{path: path}
}

func (r *FileIntentsParseRecorder) RecordIntentsParse(
	ctx context.Context,
	query string,
	intents json.RawMessage,
	prompt string,
	model string,
	temperature param.Opt[float64],
) (err error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	var record = struct {
		Query   string             `json:"query"`
		Intents json.RawMessage    `json:"intents"`
		Prompt  string             `json:"prompt"`
		Model   string             `json:"model"`
		Temp    param.Opt[float64] `json:"temperature"`
	}{
		Query:   query,
		Intents: intents,
		Prompt:  prompt,
		Model:   model,
		Temp:    temperature,
	}

	var data []byte
	if data, err = json.Marshal(&record); err != nil {
		err = fmt.Errorf("recording intents parse: marshaling record: %w", err)
		return
	}
	data = append(data, '\n')

	var f *os.File
	if f, err = os.OpenFile(
		r.path,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0o644,
	); err != nil {
		err = fmt.Errorf(
			"recording intents parse: opening file `%s`: %w",
			r.path,
			err,
		)
		return
	}
	defer f.Close()

	if _, err = f.Write(data); err != nil {
		err = fmt.Errorf(
			"recording intents parse: writing to file `%s`: %w",
			r.path,
			err,
		)
	}
	return
}
