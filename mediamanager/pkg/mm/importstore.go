package mm

import (
	"context"
	"errors"
	"fmt"
)

type ImportStore interface {
	ListImports(ctx context.Context) ([]Import, error)
	CreateImport(ctx context.Context, imp *Import) error
	DeleteImport(ctx context.Context, imp ImportID) error
	UpdateImport(ctx context.Context, imp *Import) error
}

type ImportNotFoundErr struct {
	Import ImportID `json:"import"`
}

func As[T error](err error) (typedErr T) {
	errors.As(err, &typedErr)
	return
}

func (err *ImportNotFoundErr) Error() string {
	return fmt.Sprintf("import not found: %s", err.Import)
}

type ImportExistsErr struct {
	Import ImportID `json:"import"`
}

func (err *ImportExistsErr) Error() string {
	return fmt.Sprintf("import exists: %s", err.Import)
}
