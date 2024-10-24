package mm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type ImportController struct {
	Downloads DownloadStore
	Imports   ImportStore
	Logger    *slog.Logger
	Importer  Importer
}

func (c *ImportController) Run(
	ctx context.Context,
	interval time.Duration,
) error {
	c.Logger.Info("starting import controller")

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for ; ; <-ticker.C {
		if err := c.RunLoop(ctx); err != nil {
			if errors.Is(err, context.Canceled) ||
				errors.Is(err, context.DeadlineExceeded) {
				return err
			}
			c.Logger.Error("running control loop", "err", err.Error())
		}
	}
}

func (c *ImportController) RunLoop(ctx context.Context) error {
	downloads, err := c.Downloads.ListDownloads(ctx)
	if err != nil {
		return fmt.Errorf("listing downloads: %w", err)
	}

	imports, err := c.Imports.ListImports(ctx)
	if err != nil {
		return fmt.Errorf("listing imports: %w", err)
	}

	importsByInfoHash := map[InfoHash]*Import{}
	for i := range imports {
		importsByInfoHash[imports[i].InfoHash] = &imports[i]
	}

	filteredDownloads := map[InfoHash]*Download{}
	for i := range downloads {
		if _, exists := importsByInfoHash[downloads[i].ID]; exists {
			filteredDownloads[downloads[i].ID] = &downloads[i]
		}
	}

	var wg sync.WaitGroup
	for infoHash, imp := range importsByInfoHash {
		c.reconcile(ctx, &wg, imp, filteredDownloads[infoHash])
	}
	wg.Wait()
	return nil
}

func (c *ImportController) reconcile(
	ctx context.Context,
	wg *sync.WaitGroup,
	imp *Import,
	download *Download,
) {
	logger := c.Logger.With("import", imp.ID)
	if download == nil {
		logger.Error("missing download for import", "infoHash", imp.InfoHash)
		wg.Add(1)
		go func() {
			imp.Status = ImportStatusError
			if err := c.Imports.UpdateImport(ctx, imp); err != nil {
				logger.Error("updating import", "err", err.Error())
			}
			wg.Done()
		}()
		return
	}
	logger = logger.With("infoHash", download.ID)

	// if the download is still fetching the metadata, then we just need to make
	// sure the import status is `PENDING`.
	if download.Status == DownloadStatusMetadata {
		// only do an update if the import status is not `PENDING` or `ERROR`
		if imp.Status != ImportStatusPending &&
			imp.Status != ImportStatusError {

			wg.Add(1)
			go func() {
				imp.Status = ImportStatusPending
				imp.Files = nil
				if err := c.Imports.UpdateImport(ctx, imp); err != nil {
					logger.Error(
						"updating import status",
						"err", err.Error(),
						"status", imp.Status,
					)
				}
				wg.Done()
			}()
		}
		return
	}

	if imp.Film != nil && imp.Status != ImportStatusComplete {
		clone := *imp
		clone.Status, clone.Files = c.importFilm(
			logger,
			download,
			imp.Film,
			imp.Files,
		)

		wg.Add(1)
		go func() {
			if err := c.Imports.UpdateImport(ctx, &clone); err != nil {
				logger.Error("updating import", "err", err.Error())
			}
			wg.Done()
		}()
	}
}

func (c *ImportController) importFilm(
	logger *slog.Logger,
	download *Download,
	film *Film,
	importFiles ImportFiles,
) (status ImportStatus, files ImportFiles) {
	files = make(ImportFiles, len(importFiles))
	copy(files, importFiles)
	downloadFiles := download.Files.ToMap()

	complete := c.importFile(
		logger,
		downloadFiles,
		&files,
		&LibraryFile{
			Type:        LibraryTypeFilm,
			Path:        film.PrimaryVideoFile,
			InfoHash:    download.ID,
			LibraryItem: &FilmPrimaryVideo{Title: film.Title, Year: film.Year},
		},
		&status,
	)

	for i := range film.PrimarySubtitles {
		complete = complete && c.importFile(
			logger,
			downloadFiles,
			&files,
			&LibraryFile{
				Type: LibraryTypeFilm,
				Path: film.PrimarySubtitles[i].Path,
				LibraryItem: &FilmPrimaryVideoSubtitle{
					FilmPrimaryVideo: FilmPrimaryVideo{
						Title: film.Title,
						Year:  film.Year,
					},
					Language: film.PrimarySubtitles[i].Language,
				},
			},
			&status,
		)
	}

	if status == "" && complete {
		status = ImportStatusComplete
	}
	return
}

func (c *ImportController) importFile(
	logger *slog.Logger,
	downloadFiles map[string]*DownloadFile,
	files *ImportFiles,
	file *LibraryFile,
	status *ImportStatus,
) bool {
	imfile := files.FromPathDefault(file.Path)
	dlfile := downloadFiles[file.Path]
	if dlfile == nil {
		logger.Error(
			"importing film",
			"err", "import file not found in download",
		)
		*status = ImportStatusError
	} else if imfile.Status != ImportFileStatusComplete && dlfile.Complete() {
		// if we got here, then the file has not been imported (the import file
		// does not have a `COMPLETE` status) but the file has finished
		// downloading, which suggests the file has downloaded since the last
		// time the control loop ran. we should import it.
		if err := c.Importer.ImportFile(file); err != nil {
			logger.Error(
				"importing file",
				"err", err.Error(),
				"path", file.Path,
				"libraryType", file.Type,
			)
			return false
		}
		logger.Info(
			"imported file",
			"path", file.Path,
			"libraryType", file.Type,
		)
		imfile.Status = ImportFileStatusComplete
		return true
	}
	return false
}
