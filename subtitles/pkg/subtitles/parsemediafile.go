package subtitles

import (
	"fmt"
	"io/fs"
	"regexp"
	"strings"
)

func ParseShowMediaFile(
	fsys fs.FS,
	directory string,
	fileName string,
) (epi, lng string, sub bool, err error) {
	matches := showMediaFileNameRegex.FindStringSubmatch(fileName)
	if len(matches) < 1 {
		err = fmt.Errorf(
			"parsing show media file: does not match regex: %s",
			fileName,
		)
		return
	}

	epi = matches[showMediaFileNameRegexEpisodeIndex]
	ext := matches[showMediaFileNameRegexExtIndex]
	if sub = strings.HasSuffix(ext, "srt"); sub {
		lng = ext[:strings.Index(ext, ".")]
	}

	return
}

var showMediaFileNameRegex = regexp.MustCompile(
	`^Episode (?P<episode>[0-9]{2})\.(?P<ext>[a-z]{2}\.srt|mkv|mp4)$`,
)

const (
	showMediaFileNameRegexEpisodeIndex = 1
	showMediaFileNameRegexExtIndex     = 2
)
