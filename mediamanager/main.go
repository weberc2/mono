package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-git/go-billy/v5/osfs"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	controller := TransformationController{
		transformationStore{
			"test": Transformation{
				ID:     "test",
				Status: TransformationStatusPending,
				Spec: TransformationSpec{
					Type:       TransformationTypeLink,
					SourcePath: "/foo/input",
					TargetPath: "/bar/output",
				},
			},
		},
		osfs.New("/tmp/test"),
	}

	for range time.Tick(1 * time.Second) {
		controller.runLoop(context.Background())
	}
}

type transformationStore map[TransformationID]Transformation

var _ TransformationStore = transformationStore(nil)

func (store transformationStore) UpdateTransformation(
	ctx Context,
	transformation *Transformation,
) (err error) {
	if _, exists := store[transformation.ID]; !exists {
		err = fmt.Errorf("transformation not found: %s", transformation.ID)
		return
	}
	store[transformation.ID] = *transformation
	return
}

func (store transformationStore) ListTransformations(
	ctx Context,
) (transformations []Transformation, err error) {
	for _, t := range store {
		transformations = append(transformations, t)
	}
	return
}
