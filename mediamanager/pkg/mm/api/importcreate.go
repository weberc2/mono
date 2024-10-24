package api

import (
	"context"
	"mediamanager/pkg/mm"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func (api *API) ImportCreate(
	ctx context.Context,
	input *ImportCreateInput,
) (output *ImportCreateOutput, err error) {
	output = &ImportCreateOutput{
		Body: struct {
			Import mm.Import `json:"import"`
		}{
			Import: mm.Import{
				InfoHash: input.Body.InfoHash,
				ID:       input.Import,
				Film:     input.Body.Film,
				Status:   mm.ImportStatusPending,
			},
		},
	}

	err = api.Imports.CreateImport(ctx, &output.Body.Import)

	// if there were no errors, prepare the output for serialization
	if input.Body.Film != nil {
		output.Body.Import.Files = mm.ImportFiles{
			mm.ImportFile{
				Path:   input.Body.Film.PrimaryVideoFile,
				Status: mm.ImportFileStatusPending,
			},
		}
		for i := range input.Body.Film.PrimarySubtitles {
			output.Body.Import.Files = append(
				output.Body.Import.Files,
				mm.ImportFile{
					Path:   input.Body.Film.PrimarySubtitles[i].Path,
					Status: mm.ImportFileStatusPending,
				},
			)
		}
	}
	return
}

type ImportCreateInput struct {
	Import mm.ImportID `path:"import"`
	Body   struct {
		InfoHash mm.InfoHash `json:"infoHash"`
		Film     *mm.Film    `json:"film"`
	} `required:"true"`
}

type ImportCreateOutput struct {
	Body struct {
		Import mm.Import `json:"import"`
	} `required:"true"`
}

var OperationImportCreate = Operation[ImportCreateInput, ImportCreateOutput]{
	Huma: huma.Operation{
		OperationID:   "import-create",
		Summary:       "Create import",
		Tags:          []string{"Import"},
		Path:          "/imports/{import}",
		Method:        http.MethodPost,
		DefaultStatus: http.StatusCreated,
		Errors: []int{
			http.StatusConflict, // ImportExistsErr
		},
	},
	Handler: (*API).ImportCreate,
}
