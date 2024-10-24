package api

import (
	"context"
	"mediamanager/pkg/mm"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func (api *API) ImportDelete(
	ctx context.Context,
	input *ImportDeleteInput,
) (output *ImportDeleteOutput, err error) {
	output = new(ImportDeleteOutput)
	err = api.Imports.DeleteImport(ctx, input.Import)
	return
}

type ImportDeleteInput struct {
	Import mm.ImportID `path:"import"`
}

type ImportDeleteOutput struct{}

var OperationImportDelete = Operation[ImportDeleteInput, ImportDeleteOutput]{
	Huma: huma.Operation{
		OperationID:   "import-delete",
		Summary:       "Delete import",
		Tags:          []string{"Import"},
		Path:          "/imports/{import}",
		Method:        http.MethodDelete,
		DefaultStatus: http.StatusOK,
		Errors: []int{
			http.StatusNotFound, // ImportNotFoundErr
		},
	},
	Handler: (*API).ImportDelete,
}
