package api

import (
	"context"
	"mediamanager/pkg/mm"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func (api *API) DownloadCreate(
	ctx context.Context,
	input *DownloadCreateInput,
) (output *DownloadCreateOutput, err error) {
	output = &DownloadCreateOutput{
		Body: struct {
			Download mm.Download "json:\"download\""
		}{
			Download: mm.Download{
				ID:     mm.NewInfoHash(input.InfoHash),
				Status: mm.DownloadStatusPending,
			},
		},
	}
	err = api.Downloads.CreateDownload(ctx, &output.Body.Download)
	return
}

type DownloadCreateInput struct {
	InfoHash string `path:"infoHash"`
}

type DownloadCreateOutput struct {
	Body struct {
		Download mm.Download `json:"download"`
	}
}

var OperationDownloadCreate = Operation[DownloadCreateInput, DownloadCreateOutput]{
	Huma: huma.Operation{
		OperationID:   "download-create",
		Summary:       "Create download",
		Tags:          []string{"Download"},
		Path:          "/downloads/{infoHash}",
		Method:        http.MethodPost,
		DefaultStatus: http.StatusCreated,
		Errors: []int{
			http.StatusConflict, // DownloadExistsErr
		},
	},
	Handler: (*API).DownloadCreate,
}
