package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/cenkalti/rain/rainrpc"
)

type DownloadController struct {
	DownloadDirectory string
	FilmsDirectory    string
	ShowsDirectory    string
	Downloads         DownloadStore
	Torrents          *rainrpc.Client
}

func (controller *DownloadController) controlLoop(ctx Context) {
	// TODO:
	// * handle the case where a download is deleted but the torrent remains
	//   * presumably we want to stop and delete the torrent
	//   * presently we can't tell whether a torrent is created by this
	//     controller or by some other `rain` client
	downloads, err := controller.Downloads.ListDownloads(ctx)
	if err != nil {
		slog.Error(
			"listing downloads",
			"err", err.Error(),
			"component", "DownloadController",
		)
		return
	}

	result, err := controller.Torrents.ListTorrents()
	if err != nil {
		slog.Error(
			"listing torrents",
			"err", err.Error(),
			"component", "DownloadController",
		)
		return
	}

	torrents := make(map[string]struct{}, len(result))
	for i := range result {
		torrents[result[i].ID] = struct{}{}
	}

	for i := range downloads {
		// if the download is completed, skip it
		if downloads[i].Status == DownloadStatusSuccess {
			continue
		} else if downloads[i].Status == DownloadStatusPostProcessing {
			controller.postProcessDownload(ctx, &downloads[i])
			continue
		}

		// otherwise, check if there is a corresponding torrent--if not, then
		// create one, otherwise update the download based on its torrent's
		// status.
		torrent := string(downloads[i].ID)
		if _, exists := torrents[torrent]; !exists {
			controller.createTorrentForDownload(ctx, &downloads[i])
		} else {
			controller.updateDownloadFromTorrent(ctx, downloads[i].ID, torrent)
		}
	}
}

func (controller *DownloadController) postProcessDownload(
	ctx Context,
	download *Download,
) {
	if err := ctx.Err(); err != nil {
		slog.Error(
			"aborting post-processing: context canceled or deadline expired",
			"err", err.Error(),
			"component", "DownloadController",
			"download", download.ID,
		)
		return
	}

	// process the first N-1 files, keeping track of failures
	var errorMessages []string
	for i := range download.Spec.Files.List {
		if err := controller.processFile(
			ctx,
			download,
			&download.Spec.Files.List[i],
		); err != "" {
			errorMessages = append(errorMessages, err)
		} else {
			download.ProcessedFiles = append(
				download.ProcessedFiles,
				download.Spec.Files.List[i].Path,
			)
		}
	}

	// if all files have been processed successfully, the status should be
	// `SUCCESS`, otherwise `POST_PROCESSING`
	var status DownloadStatus
	if len(errorMessages) > 0 {
		status = DownloadStatusPostProcessing
	} else {
		status = DownloadStatusSuccess
	}

	controller.updateDownloadStatus(
		ctx,
		download.ID,
		status,
		download.Torrent,
		download.ProcessedFiles,
		strings.Join(errorMessages, "\n"),
	)
}

func (controller *DownloadController) processFile(
	ctx Context,
	download *Download,
	file *MediaFile,
) (errorMessage string) {
	// check to see if the file has already been processed; if so, skip it
	for i := range download.ProcessedFiles {
		if download.ProcessedFiles[i] == file.Path {
			return
		}
	}

	ext := filepath.Ext(file.Path)
	if file.Kind == FileKindSubtitle && file.Lang != "" {
		ext = fmt.Sprintf(".%s%s", file.Lang, ext)
	}
	var linkFile string
	parent := fmt.Sprintf("%s (%s)", file.Title, file.Year)
	switch file.Type {
	case MediaTypeFilm:
		linkFile = filepath.Join(
			controller.FilmsDirectory,
			parent,
			file.Title+ext,
		)
	case MediaTypeEpisode:
		linkFile = filepath.Join(
			controller.ShowsDirectory,
			parent,
			fmt.Sprintf("Season %s", file.Season),
			fmt.Sprintf("Episode %s%s", file.Episode, ext),
		)
	}

	linkFileTarget := filepath.Join(
		controller.DownloadDirectory,
		string(download.ID),
		file.Path,
	)
	if err := os.MkdirAll(filepath.Dir(linkFile), 0755); err != nil {
		slog.Error(
			"linking downloaded file: creating parent directories",
			"err", err.Error(),
			"component", "DownloadController",
			"download", download.ID,
			"linkFile", linkFile,
			"linkFileTarget", linkFileTarget,
		)
		errorMessage = fmt.Sprintf(
			"linking downloaded file: creating parent directories: %v",
			err,
		)
	} else if err := os.Link(
		linkFileTarget,
		linkFile,
	); err != nil && !os.IsExist(err) {
		slog.Error(
			"linking downloaded file",
			"err", err.Error(),
			"component", "DownloadController",
			"download", download.ID,
			"linkFile", linkFile,
			"linkFileTarget", linkFileTarget,
		)
		errorMessage = fmt.Sprintf("linking downloaded file: %v", err)
	} else {
		slog.Info(
			"linked downloaded file",
			"download", download.ID,
			"linkFileTarget", linkFileTarget,
			"linkFile", linkFile,
		)
	}
	return
}

func (controller *DownloadController) handlePending(
	ctx Context,
	download *Download,
) {
	// if the download is PENDING but there's already a torrent
	// corresponding to the download, then the download should be
	// updated according to the torrent's status.
	torrent := string(download.ID)
	controller.updateDownloadFromTorrent(ctx, download.ID, torrent)
}

func (controller *DownloadController) updateDownloadFromTorrent(
	ctx Context,
	download DownloadID,
	torrent string,
) {
	stats, err := controller.Torrents.GetTorrentStats(torrent)
	if err != nil {
		slog.Error(
			"fetching status for torrent",
			"err", err.Error(),
			"component", "DownloadController",
			"torrent", torrent,
			"download", download,
		)
		return
	}

	var status DownloadStatus
	if stats.Bytes.Completed == stats.Bytes.Total {
		status = DownloadStatusPostProcessing
	} else if stats.Status == "Downloading Metadata" {
		status = DownloadStatusFetchingMetadata
	} else if stats.Status == "Verifying" {
		status = DownloadStatusVerifying
	} else {
		status = DownloadStatusDownloading
	}

	controller.updateDownloadStatus(
		ctx,
		download,
		status,
		torrent,
		nil,
		stats.Error,
	)
}

func (controller *DownloadController) createTorrentForDownload(
	ctx Context,
	download *Download,
) {
	var errorMessage string
	var torrent string
	var status DownloadStatus

	if err := download.Spec.Source.addToRain(
		controller.Torrents,
		&rainrpc.AddTorrentOptions{ID: string(download.ID)},
	); err != nil {
		errorMessage = fmt.Sprintf("creating torrent for download: %v", err)
		torrent = download.Torrent
		status = download.Status
		slog.Error(
			"creating torrent for download",
			"err", err.Error(),
			"component", "DownloadController",
			"download", download.ID,
			"downloadType", download.Spec.Source.Type,
		)
	} else {
		torrent = string(download.ID)
		status = DownloadStatusFetchingMetadata
	}

	controller.updateDownloadStatus(
		ctx,
		download.ID,
		status,
		torrent,
		nil,
		errorMessage,
	)
}

func (controller *DownloadController) updateDownloadStatus(
	ctx Context,
	download DownloadID,
	status DownloadStatus,
	torrent string,
	processedFiles []string,
	errorMessage string,
) {
	if _, err := controller.Downloads.UpdateDownloadStatus(
		ctx,
		download,
		status,
		torrent,
		processedFiles,
		errorMessage,
	); err != nil {
		slog.Error(
			"updating download status",
			"err", err.Error(),
			"component", "DownloadController",
			"download", download,
			"status", status,
			"torrent", torrent,
			"processedFiles", processedFiles,
			"errorMessage", errorMessage,
		)
	}
}

type Context = context.Context
