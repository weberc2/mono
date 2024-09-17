package main

import (
	"fmt"
	"os"
	"path/filepath"
)

type Film struct {
	Title string
	Year  string
}

func (f *Film) String() string {
	return fmt.Sprintf("%s (%s)", f.Title, f.Year)
}

type FilmMediaFile struct {
	Film
	SourcePath string
}

func (file *FilmMediaFile) NormalizedFileName() string {
	return file.Film.String() + filepath.Ext(file.SourcePath)
}

type Linker struct {
	DownloadDirectory string
	FilmsDirectory    string
}

func (l *Linker) LinkFilm(mediaFile *FilmMediaFile) error {
	dir := filepath.Join(l.FilmsDirectory, mediaFile.Film.String())
	dst := filepath.Join(dir, mediaFile.NormalizedFileName())
	src := filepath.Join(l.DownloadDirectory, mediaFile.SourcePath)
	if err := os.Mkdir(dir, 0755); err != nil {
		return fmt.Errorf(
			"linking film `%s`: creating film directory: %w",
			mediaFile.Film.String(),
			err,
		)
	}
	if err := os.Link(src, dst); err != nil {
		return fmt.Errorf("linking film `%s`: %w", mediaFile.Film.String(), err)
	}
	return nil
}
