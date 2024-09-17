package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/go-git/go-billy/v5"
)

type TransformationController struct {
	Transformations TransformationStore
	Files           billy.Filesystem
}

func (controller *TransformationController) runLoop(ctx Context) {
	transformations, err := controller.Transformations.ListTransformations(ctx)
	if err != nil {
		slog.Error(
			"listing transformations",
			"err", err.Error(),
			"controller", "TransformationController",
		)
		return
	}

	for i := range transformations {
		controller.handleTransformation(ctx, &transformations[i])
	}
}

func (controller *TransformationController) handleTransformation(
	ctx Context,
	transformation *Transformation,
) {
	switch transformation.Status {
	case TransformationStatusPending, TransformationStatusError:
		slog.Debug(
			"handling transformation",
			"controller", "TransformationController",
			"transformation", transformation.ID,
			"status", transformation.Status,
		)
		break
	case TransformationStatusSuccess, TransformationStatusFailure:
		// nothing to do; already in terminal state
		slog.Debug(
			"transformation completed",
			"controller", "TransformationController",
			"transformation", transformation.ID,
			"status", transformation.Status,
		)
		return
	default:
		slog.Error(
			"invalid transformation status",
			"controller", "TransformationController",
			"status", transformation.Status,
			"transformation", transformation.ID,
		)
		return
	}

	if _, err := controller.Files.Stat(
		transformation.Spec.TargetPath,
	); err != nil {
		if !os.IsNotExist(err) {
			slog.Error(
				"handling pending transformation: checking for target file",
				"err", err.Error(),
				"controller", "TransformationController",
				"transformation", transformation.ID,
				"targetPath", transformation.Spec.TargetPath,
			)
			return
		}

		// if the file doesn't exist, that's expected--just apply the
		// transformation
		switch transformation.Spec.Type {
		case TransformationTypeLink:
			if err := controller.Files.Symlink(
				transformation.Spec.SourcePath,
				transformation.Spec.TargetPath,
			); err != nil {
				slog.Error(
					"handling pending transformation: failed to link target "+
						"file",
					"err", err.Error(),
					"controller", "TransformationController",
					"transformation", transformation.ID,
					"sourcePath", transformation.Spec.SourcePath,
					"targetPath", transformation.Spec.TargetPath,
				)

				clone := *transformation
				clone.Error = fmt.Sprintf("failed to link target: %v", err)
				clone.Status = TransformationStatusError

				if err = controller.Transformations.UpdateTransformation(
					ctx,
					&clone,
				); err != nil {
					slog.Error(
						"handling pending transformation: "+
							"handling target link error",
						"err", err.Error(),
						"controller", "TransformationController",
						"transformation", transformation.ID,
					)
				}
			}
		default:
			clone := *transformation
			clone.Error = fmt.Sprintf(
				"unsupported transformation type: %s",
				transformation.Spec.Type,
			)

			slog.Error(
				"handling pending transformation: unsupported type",
				"err", clone.Error,
				"controller", "TransformationController",
				"transformation", transformation.ID,
			)

			if err := controller.Transformations.UpdateTransformation(
				ctx,
				&clone,
			); err != nil {
				slog.Error(
					"handling pending transformation: "+
						"marking transformation failed",
					"err", err.Error(),
					"controller", "TransformationController",
					"transformation", transformation.ID,
				)
			}
			return
		}
	}

	// if the file exists, then we'll assume that the transformation succeeded.
	// update the transformation store and return.
	clone := *transformation
	clone.Error = ""
	clone.Status = TransformationStatusSuccess
	if err := controller.Transformations.UpdateTransformation(
		ctx,
		&clone,
	); err != nil {
		slog.Error(
			"handling pending transformation: marking success",
			"err", err.Error(),
			"controller", "TransformationController",
			"transformation", transformation.ID,
		)
	}
}

type TransformationStore interface {
	UpdateTransformation(Context, *Transformation) (err error)
	ListTransformations(Context) (transformations []Transformation, err error)
}
