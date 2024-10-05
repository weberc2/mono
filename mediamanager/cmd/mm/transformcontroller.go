package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

type TransformController struct {
	Downloads          DownloadStore
	Transforms         TransformStore
	DownloadsDirectory string
	FilmsDirectory     string
	Logger             *slog.Logger
}

func NewTransformController(transforms TransformStore) (c TransformController) {
	c.Transforms = transforms
	c.Logger = slog.With("controller", "TRANSFORM")
	return
}

func (c *TransformController) Run(ctx Context, interval time.Duration) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		c.RunLoop(ctx)
		<-ticker.C
	}
}

func (c *TransformController) RunLoop(ctx Context) {
	c.Logger.Debug("running control loop")
	transforms, err := c.Transforms.ListTransforms(ctx)
	if err != nil {
		c.Logger.Error("listing transforms", "err", err.Error())
		return
	}
	c.Logger.Debug("fetched transforms", "count", len(transforms))

	downloads, err := c.Downloads.ListDownloads(ctx)
	if err != nil {
		c.Logger.Error("listing downloads", "err", err.Error())
		return
	}
	c.Logger.Debug("fetched downloads", "count", len(downloads))

	// create an iterator over transform files that are in PENDING status but
	// whose corresponding download file has finished downloading. start by
	// taking all transforms and filtering out those transforms whose
	// corresponding download has not yet begun.
	// pendingFiles := filterFilesIter{
	// 	files: &filesIter{
	// 		transforms: transformIter{
	// 			transforms: transforms,
	// 			downloads:  downloads,
	// 		},
	// 	},
	// 	filter: func(tf *TransformFile, df *DownloadFile, err error) bool {
	// 		return tf.Status == TransformFileStatusPending &&
	// 			df.Progress >= df.Size
	// 	},
	// }

	downloadsByInfoHash := make(map[InfoHash]*Download)
	for i := range downloads {
		downloadsByInfoHash[downloads[i].InfoHash] = &downloads[i]
	}

	for i := range transforms {
		if len(transforms[i].Files) < 1 {
			c.Logger.Warn(
				"transform has no associated files: skipping",
				"transform", transforms[i].ID,
			)
			continue
		}

		download, exists := downloadsByInfoHash[transforms[i].Spec.InfoHash]
		if !exists {
			c.Logger.Warn(
				"transform has no associated download: skipping",
				"transform", transforms[i].ID,
				"infoHash", transforms[i].Spec.InfoHash,
			)
			continue
		}

		downloadFiles := make(map[string]*DownloadFile)
		for i := range download.Files {
			downloadFiles[download.Files[i].Path] = &download.Files[i]
		}

		for j := range transforms[i].Files {
			if transforms[i].Files[j].Status == TransformFileStatusSuccess {
				c.Logger.Debug(
					"transform file already successful: skipping",
					"transform", transforms[i],
					"file", transforms[i].Files[j].Path,
				)
				continue
			}
			df, exists := downloadFiles[transforms[i].Files[j].Path]
			if !exists {
				c.Logger.Warn(
					"transform file missing corresponding download file: "+
						"skipping",
					"transform", transforms[i].ID,
					"infoHash", transforms[i].Spec.InfoHash,
					"file", transforms[i].Files[j].Path,
				)
				continue
			}

			if df.Progress < df.Size {
				c.Logger.Debug(
					"download file not yet fully downloaded: skipping",
					"transform", transforms[i].ID,
					"infoHash", transforms[i].Spec.InfoHash,
					"file", df.Path,
				)
				continue
			}

			c.transformFile(ctx, &transforms[i], &transforms[i].Files[j])
		}
	}

	//for {
	//	t, tf, _, err := pendingFiles.next()
	//	if err != nil {
	//		c.Logger.Error("fetching next pending file", "err", err.Error())
	//		continue
	//	}

	//	// eof
	//	if tf == nil {
	//		break
	//	}

	//	c.transformFile(ctx, t, tf)
	//}
}

func (c *TransformController) transformFile(
	ctx Context,
	t *Transform,
	file *TransformFile,
) {
	c.Logger.Debug("transforming file", "transform", t.ID, "file", file.Path)
	var linkFile string
	switch t.Spec.Type {
	case TransformTypeFilm:
		linkFile = filepath.Join(
			c.FilmsDirectory,
			fmt.Sprintf(
				"%s (%s)/%s%s",
				t.Spec.Film.Title,
				t.Spec.Film.Year,
				t.Spec.Film.Title,
				filepath.Ext(file.Path),
			),
		)
	default:
		slog.Error(
			"transforming file",
			"err", fmt.Sprintf("unsupported transform type: %s", t.Spec.Type),
			"transformType", t.Spec.Type,
		)
		return
	}

	sourceFile := filepath.Join(
		c.DownloadsDirectory,
		string(t.Spec.InfoHash),
		file.Path,
	)

	var attempt int
	for attempt = range 3 {
		if err := os.MkdirAll(filepath.Dir(linkFile), 0766); err != nil {
			slog.Error(
				"creating parent directories for link file",
				"err", err.Error(),
				"linkFile", linkFile,
			)
			continue
		}
		if err := os.Link(sourceFile, linkFile); err != nil {
			if os.IsExist(err) {
				if err := os.Remove(linkFile); err != nil {
					slog.Error(
						"removing vestigial link file",
						"err", err.Error(),
						"linkFile", linkFile,
					)
					return
				}
				continue
			}
			slog.Error(
				"linking file",
				"err", err.Error(),
				"sourceFile", sourceFile,
				"linkFile", linkFile,
			)
			continue
		}
		goto UPDATE
	}
	slog.Error(
		"linking file: attempts exceeded limit",
		"attempts", attempt,
		"linkFile", linkFile,
	)
	return

UPDATE:
	file.Status = TransformFileStatusSuccess
	if _, err := c.Transforms.UpdateTransformFile(ctx, t.ID, file); err != nil {
		slog.Error(
			"updating transform file",
			"err", err.Error(),
			"sourceFile", sourceFile,
			"linkFile", linkFile,
		)
	}

	slog.Info(
		"successfully linked transform file",
		"sourceFile", sourceFile,
		"linkFile", linkFile,
	)
}

type TransformStore interface {
	ListTransforms(ctx Context) ([]Transform, error)
	CreateTransform(ctx Context, t *Transform) (Transform, error)
	UpdateTransformFile(
		ctx Context,
		id TransformID,
		file *TransformFile,
	) (TransformFile, error)
}

type DownloadStore interface {
	CreateDownload(ctx Context, d *Download) (Download, error)
	FetchDownload(ctx Context, id DownloadID) (Download, error)
	ListDownloads(ctx Context) ([]Download, error)
}
