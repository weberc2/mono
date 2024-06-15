package subtitles

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
)

func ParseShowMediaFile(
	fsys fs.FS,
	directory string,
	fileName string,
) (mf MediaFile, err error) {
	matches := showMediaFileNameRegex.FindStringSubmatch(fileName)
	if len(matches) < 1 {
		err = fmt.Errorf(
			"parsing show media file: does not match regex: %s",
			fileName,
		)
		return
	}

	mf.Type = MediaTypeShow
	mf.Title = matches[showMediaFileNameRegexShowIndex]
	mf.Season = matches[showMediaFileNameRegexSeasonIndex]
	mf.Episode = matches[showMediaFileNameRegexEpisodeIndex]
	mf.Filepath = filepath.Join(directory, fileName)

	if matches[showMediaFileNameRegexSrtExtIndex] == "srt" {
		mf.Kind = MediaFileKindSubtitle
		mf.SubtitleLanguage = matches[showMediaFileNameRegexLangIndex]
		// TODO: compute framerate

	} else if matches[showMediaFileNameRegexMkvExtIndex] == "mkv" {
		mf.Kind = MediaFileKindVideo
		var f fs.File
		if f, err = fsys.Open(mf.Filepath); err != nil {
			err = fmt.Errorf(
				"parsing media file for show `%s S%sE%s`: opening file: %w",
				mf.Title,
				mf.Season,
				mf.Episode,
				err,
			)
			return
		}
		defer func() {
			if closeErr := f.Close(); closeErr != nil {
				closeErr = fmt.Errorf(
					"parsing media file for show `%s S%sE%s: closing file: %w`",
					mf.Title,
					mf.Season,
					mf.Episode,
					err,
				)
				err = errors.Join(err, closeErr)
			}
		}()

		if mf.Mediahash, err = Mediahash(f); err != nil {
			err = fmt.Errorf(
				"parsing media file for show `%s S%sE%s`: "+
					"computing mediahash: %w",
				mf.Title,
				mf.Season,
				mf.Episode,
				err,
			)
			return
		}

		// TODO: compute framerate
	}
	return
}

var showMediaFileNameRegex = regexp.MustCompile(
	`^(?P<show>.*) S(?P<season>\d{2})E(?P<episode>\d{2})(?:(?:\.(?P<lang>[a-zA-Z-]{2,9})\.(?P<ext>srt))|\.(?P<ext>mkv))$`,
)

const (
	showMediaFileNameRegexShowIndex    = 1
	showMediaFileNameRegexSeasonIndex  = 2
	showMediaFileNameRegexEpisodeIndex = 3
	showMediaFileNameRegexLangIndex    = 4
	showMediaFileNameRegexSrtExtIndex  = 5
	showMediaFileNameRegexMkvExtIndex  = 6
)
