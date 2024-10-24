package api

import (
	"context"
	"mediamanager/pkg/mm"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func (api *API) ImportList(
	ctx context.Context,
	input *ImportListInput,
) (output *ImportListOutput, err error) {
	output = new(ImportListOutput)
	output.Body.Imports, err = api.Imports.ListImports(ctx)
	return
}

type ImportListInput struct{}

type ImportListOutput struct {
	Body struct {
		Imports mm.Slice[mm.Import] `json:"imports"`
	} `required:"true"`
}

var OperationImportList = Operation[ImportListInput, ImportListOutput]{
	Huma: huma.Operation{
		OperationID:   "import-list",
		Summary:       "List imports",
		Tags:          []string{"Import"},
		Path:          "/imports",
		Method:        http.MethodGet,
		DefaultStatus: http.StatusOK,
	},
	Handler: (*API).ImportList,
}
