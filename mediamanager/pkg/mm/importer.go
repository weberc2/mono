package mm

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Importer struct {
	DownloadsDirectory string
	FilmsDirectory     string
	ScratchDirectory   string
}

func (i *Importer) ImportFile(file *LibraryFile) (err error) {
	target := filepath.Join(i.libraryDirectory(file), file.LibraryPath())
	source := filepath.Join(i.DownloadsDirectory, file.DownloadPath())
	scratch := filepath.Join(i.ScratchDirectory, file.InfoHash.String())
	temp := filepath.Join(scratch, filepath.Base(target))

	if err = os.MkdirAll(scratch, 0766); err != nil {
		err = fmt.Errorf("importing film file: creating scratch dir: %w", err)
		return
	}

	if err = os.Remove(temp); err != nil && !os.IsNotExist(err) {
		err = fmt.Errorf("importing film file: %w", err)
		return
	}

	if err = os.Link(source, temp); err != nil {
		err = fmt.Errorf("importing film file: %w", err)
		return
	}

	if err = os.MkdirAll(filepath.Dir(target), 0766); err != nil {
		err = errors.Join(
			os.Remove(temp),
			fmt.Errorf("importing film file: %w", err),
		)
		return
	}

	if err = os.Rename(temp, target); err != nil {
		err = fmt.Errorf("importing film file: %w", err)
		return
	}

	return nil
}

func (i *Importer) libraryDirectory(file *LibraryFile) string {
	switch libraryType := file.Type; libraryType {
	case LibraryTypeFilm:
		return i.FilmsDirectory
	default:
		panic(fmt.Sprintf("invalid library type: %s", libraryType))
	}
}

type LibraryFile struct {
	Type        LibraryType
	Path        string
	InfoHash    InfoHash
	LibraryItem LibraryItem
}

func (file *LibraryFile) DownloadPath() string {
	return filepath.Join(file.InfoHash.String(), file.Path)
}

func (file *LibraryFile) LibraryPath() string {
	return file.LibraryItem.LibraryPath(file.Path)
}

type LibraryItem interface {
	LibraryPath(path string) string
}

type FilmPrimaryVideoSubtitle struct {
	FilmPrimaryVideo
	Language string
}

func (file *FilmPrimaryVideoSubtitle) LibraryPath(path string) string {
	return filepath.Join(
		file.FilmDirectory(),
		file.Title+"."+file.Language+filepath.Ext(path),
	)
}

type FilmPrimaryVideo struct {
	Title string
	Year  string
}

func (file *FilmPrimaryVideo) LibraryPath(path string) string {
	return filepath.Join(file.FilmDirectory(), file.Title+filepath.Ext(path))
}

func (file *FilmPrimaryVideo) FilmDirectory() string {
	return fmt.Sprintf("%s (%s)", file.Title, file.Year)
}

type LibraryType string

const (
	LibraryTypeFilm LibraryType = "FILM"
)
