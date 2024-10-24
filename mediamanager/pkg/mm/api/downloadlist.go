package api

import (
	"context"
	"mediamanager/pkg/mm"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func (api *API) DownloadList(
	ctx context.Context,
	input *DownloadListInput,
) (output *DownloadListOutput, err error) {
	output = new(DownloadListOutput)
	output.Body.Downloads, err = api.Downloads.ListDownloads(ctx)
	return
}

type DownloadListInput struct{}

type DownloadListOutput struct {
	Body struct {
		Downloads mm.Slice[mm.Download] `json:"downloads"`
	} `required:"true"`
}

var OperationDownloadList = Operation[DownloadListInput, DownloadListOutput]{
	Huma: huma.Operation{
		OperationID:   "download-list",
		Summary:       "List downloads",
		Tags:          []string{"Download"},
		Path:          "/downloads",
		Method:        http.MethodGet,
		DefaultStatus: http.StatusOK,
	},
	Handler: (*API).DownloadList,
}
