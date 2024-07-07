package subtitles

import (
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"path/filepath"
	"regexp"
)

type SeasonIterator struct {
	Directory   string
	Show        string
	Year        string
	Season      string
	Files       []fs.DirEntry
	CurrentFile int
}

func IterateSeason(
	fsys fs.FS,
	show string,
	year string,
	showDirectory string,
	seasonFileName string,
) (iter SeasonIterator, err error) {
	iter.Show = show
	iter.Year = year
	iter.Directory = filepath.Join(showDirectory, seasonFileName)

	if iter.Season, err = ParseSeasonFileName(seasonFileName); err != nil {
		err = fmt.Errorf("parsing season: %w", err)
		return
	}

	if iter.Files, err = fs.ReadDir(fsys, iter.Directory); err != nil {
		err = fmt.Errorf("parsing season `%s`: %w", iter.Season, err)
		return
	}

	return
}

func (iter *SeasonIterator) Next(fsys fs.FS) (mf MediaFile[Episode], err error) {
NEXT_FILE:
	if iter.CurrentFile < len(iter.Files) {
		file := iter.Files[iter.CurrentFile]
		if file.IsDir() {
			slog.Warn(
				"parsing season: skipping unexpected subdirectory",
				"series", iter.Show,
				"season", iter.Season,
				"subdirectory", file.Name(),
			)
			iter.CurrentFile++
			goto NEXT_FILE
		}

		fileName := file.Name()
		if mf.ID.Episode, mf.Language, mf.IsSubtitle, err = ParseShowMediaFile(
			fsys,
			iter.Directory,
			fileName,
		); err != nil {
			err = fmt.Errorf("parsing season `%s`: %w", iter.Season, err)
			return
		}
		mf.ID.Title = iter.Show
		mf.ID.Year = iter.Year
		mf.ID.Season = iter.Season
		mf.Filepath = filepath.Join(iter.Directory, fileName)

		iter.CurrentFile++
		return
	}

	// if we've iterated through all of the seasons, return `io.EOF`
	err = io.EOF
	return
}

func ParseSeasonFileName(fileName string) (season string, err error) {
	matches := seasonRegex.FindStringSubmatch(fileName)
	if len(matches) < 1 {
		err = fmt.Errorf(
			"parsing season filename: does not match regex: %s",
			fileName,
		)
		return
	}

	season = matches[seasonRegexSeasonIndex]
	return
}

var seasonRegex = regexp.MustCompile(`^Season (?P<season>[0-9]{2})$`)

const seasonRegexSeasonIndex = 1
