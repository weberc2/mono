package subtitles

import (
	"fmt"
	"strings"
)

type Episode struct {
	Title   string `json:"title"`
	Year    string `json:"year"`
	Season  string `json:"season"`
	Episode string `json:"episode"`
}

func (e Episode) Debug(sb *strings.Builder) {
	fmt.Fprintf(
		sb,
		"title=%s, year=%s, season=%s, episode=%s",
		e.Title,
		e.Year,
		e.Season,
		e.Episode,
	)
}
