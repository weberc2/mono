package api

import (
	"context"
	"mediamanager/pkg/mm"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func (api *API) DownloadFetch(
	ctx context.Context,
	input *DownloadFetchInput,
) (output *DownloadFetchOutput, err error) {
	output = new(DownloadFetchOutput)
	output.Body.Download, err = api.Downloads.FetchDownload(
		ctx,
		mm.NewInfoHash(input.InfoHash),
	)
	return
}

type DownloadFetchInput struct {
	InfoHash string `path:"infoHash"`
}

type DownloadFetchOutput struct {
	Body struct {
		Download mm.Download `json:"download"`
	}
}

var OperationDownloadFetch = Operation[DownloadFetchInput, DownloadFetchOutput]{
	Huma: huma.Operation{
		OperationID:   "download-fetch",
		Summary:       "Fetch download",
		Tags:          []string{"Download"},
		Path:          "/downloads/{infoHash}",
		Method:        http.MethodGet,
		DefaultStatus: http.StatusOK,
		Errors:        []int{http.StatusNotFound}, // DownloadNotFoundErr
	},
	Handler: (*API).DownloadFetch,
}
