package main

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
)

type MediaDownloadController struct {
	DownloadDirectory string
	MediaDirectory    string
	MediaDownloads    MediaDownloadStore
	Downloader        Downloader
}

func (controller *MediaDownloadController) controlLoop(ctx Context) error {
	panic("not implemented")
	// mediaDownloads, err := controller.MediaDownloads.ListMediaDownloads(ctx)
	// if err != nil {
	// 	return err
	// }

	// downloads, err := controller.Downloader.ListDownloads(ctx)
	// if err != nil {
	// 	return err
	// }

	// downloadsByMediaDownload := map[MediaDownloadID]*Download{}

	// for i := range mediaDownloads {
	// 	if mediaDownloads[i].Download == "" {
	// 		controller.submitDownload(ctx, &mediaDownloads[i])
	// 	}
	// }
}

func (controller *MediaDownloadController) submitDownload(
	ctx Context,
	mediaDownload *MediaDownload,
) {
	panic("not implemented")
	// download, err := controller.Downloader.StartDownload(
	// 	ctx,
	// 	&Download{
	// 		ID: DownloadID(mediaDownload.ID),
	// 		Annotations: map[string]string{
	// 			"weberc2.com/mediamanager/mediadownload": string(mediaDownload.ID),
	// 		},
	// 		Spec: mediaDownload.Spec.Download,
	// 	},
	// )
	// if err != nil {
	// 	slog.Error(
	// 		"submitting download",
	// 		"err", err.Error(),
	// 		"controller", "MediaDownloadController",
	// 		"mediaDownload", mediaDownload.ID,
	// 	)
	// }
}

func validateMediaDownload(spec *MediaDownloadSpec) (err error) {
	panic("not implemented")
	// if spec.Files.Type == MediaFilesTypeList {
	// 	var metaInfoPtr *metainfo.MetaInfo
	// 	switch spec.DownloadSpec.Type {
	// 	case MediaSourceTypeTorrent:
	// 		var metaInfo metainfo.MetaInfo
	// 		metaInfoPtr = &metaInfo
	// 		if err = bencode.Unmarshal(
	// 			[]byte(spec.DownloadSpec.Torrent),
	// 			metaInfoPtr,
	// 		); err != nil {
	// 			err = fmt.Errorf(
	// 				"creating media download: unmarshaling torrent: %w",
	// 				err,
	// 			)
	// 			return
	// 		}
	// 	case MediaSourceTypeMetaInfo:
	// 		metaInfoPtr = &spec.DownloadSpec.MetaInfo
	// 	default:
	// 		return
	// 	}

	// 	if err = validateFileList(metaInfoPtr, spec.Files.List); err != nil {
	// 		err = fmt.Errorf("validating media download: %w", err)
	// 		return
	// 	}
	// }
	// return
}

func validateFileList(
	metaInfo *metainfo.MetaInfo,
	files []MediaFile,
) (err error) {
	var info metainfo.Info
	if err = bencode.Unmarshal(metaInfo.InfoBytes, &info); err != nil {
		err = fmt.Errorf(
			"validating file list: unmarshaling torrent `info`: %w",
			err,
		)
		return
	}

OUTER:
	for i := range files {
		cleaned := filepath.Clean(files[i].Path)
		for j := range info.Files {
			joined := filepath.Join(info.Files[j].Path...)
			if cleaned == joined {
				continue OUTER
			}
		}

		err = fmt.Errorf(
			"validating file list: file not found in source: %s",
			cleaned,
		)
		return
	}

	return nil
}

type Downloader interface {
	ListDownloads(Context) (downloads []Download, err error)
	StartDownload(Context, *Download) (download Download, err error)
}

type MediaDownloadStore interface {
	ListMediaDownloads(Context) (downloads []MediaDownload, err error)
	CreateMediaDownload(Context, *MediaDownload) (download MediaDownload, err error)
}

var ErrMediaDownloadExists = errors.New("download exists")

type Context = context.Context
