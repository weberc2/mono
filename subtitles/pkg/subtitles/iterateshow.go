package subtitles

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"path/filepath"
	"regexp"
)

type ShowIterator struct {
	Directory     string
	Title         string
	Year          string
	Files         []fs.DirEntry
	CurrentFile   int
	CurrentSeason SeasonIterator
}

func IterateShow(
	fsys fs.FS,
	showsDirectory string,
	showFileName string,
) (it ShowIterator, err error) {
	if it.Title, it.Year, err = ParseShowFileName(showFileName); err != nil {
		err = fmt.Errorf("iterating over show directory: %w", err)
		return
	}
	it.Directory = filepath.Join(showsDirectory, showFileName)
	if it.Files, err = fs.ReadDir(fsys, it.Directory); err != nil {
		err = fmt.Errorf(
			"iterating over series `%s (%s)`: %w",
			it.Title,
			it.Year,
			err,
		)
	}
	return
}

func (iter *ShowIterator) Next(fsys fs.FS) (mf MediaFile[Episode], err error) {
TOP:
	// try getting the next media file from the current season iterator. if no
	// error is returned, return the media file immediately.
	if mf, err = iter.CurrentSeason.Next(fsys); err == nil {
		return
	}

	// if there is an error and it's not `io.EOF` then return the error.
	if !errors.Is(err, io.EOF) {
		err = fmt.Errorf(
			"parsing series `%s (%s)`: %w",
			iter.Title,
			iter.Year,
			err,
		)
		return
	}

	// if the error was `io.EOF` it means we need to start iterating over the
	// next season.
NEXT_FILE:
	if iter.CurrentFile < len(iter.Files) {
		file := iter.Files[iter.CurrentFile]
		if !file.IsDir() {
			slog.Warn(
				"parsing series directory: skipping unexpected file",
				"series", iter.Title,
				"seriesDirectory", iter.Directory,
				"fileName", file.Name(),
				"fileType", file.Type().String(),
			)
			iter.CurrentFile++
			goto NEXT_FILE
		}

		if iter.CurrentSeason, err = IterateSeason(
			fsys,
			iter.Title,
			iter.Year,
			iter.Directory,
			file.Name(),
		); err != nil {
			err = fmt.Errorf(
				"parsing series `%s (%s)`: %w",
				iter.Title,
				iter.Year,
				err,
			)
			return
		}

		iter.CurrentFile++
		goto TOP
	}

	// if we've iterated through all of the seasons, return `io.EOF`
	err = io.EOF
	return
}

func ParseShowFileName(fileName string) (title, year string, err error) {
	matches := seriesRegex.FindStringSubmatch(fileName)
	if len(matches) < 1 {
		err = fmt.Errorf(
			"parsing series filename: does not match regex: %s",
			fileName,
		)
		return
	}
	title = matches[seriesRegexTitleIndex]
	year = matches[seriesRegexYearIndex]
	return
}

var seriesRegex = regexp.MustCompile(
	`^(?P<title>.+) \((?P<year>[0-9]{4})\)$`,
)

const (
	seriesRegexTitleIndex = 1
	seriesRegexYearIndex  = 2
)
