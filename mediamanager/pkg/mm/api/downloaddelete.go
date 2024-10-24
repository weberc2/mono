package api

import (
	"context"
	"mediamanager/pkg/mm"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func (api *API) DownloadDelete(
	ctx context.Context,
	input *DownloadDeleteInput,
) (output *DownloadDeleteOutput, err error) {
	// TODO: should deleting the download actually delete the record and orphan
	// the downloaded files? or should it signal to the controller somehow that
	// the download should be cleaned up first? maybe add `DELETING` and
	// `DELETED` statuses?
	output = new(DownloadDeleteOutput)
	err = api.Downloads.DeleteDownload(ctx, mm.NewInfoHash(input.InfoHash))
	return
}

type DownloadDeleteInput struct {
	InfoHash string `path:"infoHash"`
}

type DownloadDeleteOutput struct{}

var OperationDownloadDelete = Operation[DownloadDeleteInput, DownloadDeleteOutput]{
	Huma: huma.Operation{
		OperationID:   "download-delete",
		Summary:       "Delete download",
		Tags:          []string{"Download"},
		Path:          "/downloads/{infoHash}",
		Method:        http.MethodDelete,
		DefaultStatus: http.StatusOK,
	},
	Handler: (*API).DownloadDelete,
}
