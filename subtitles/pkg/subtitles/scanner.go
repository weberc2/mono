package subtitles

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
)

type Scanner struct {
	Model          Model
	FileSystem     fs.FS
	ShowsDirectory string
}

func (s *Scanner) Scan(ctx context.Context) error {
	shows, err := IterateShows(s.FileSystem, s.ShowsDirectory)
	if err != nil {
		return fmt.Errorf("scanning media: %w", err)
	}

	for {
		file, err := shows.Next(s.FileSystem)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return fmt.Errorf("scanning media: %w", err)
		}

		if file.IsSubtitle {
			err = s.Model.InsertEpisodeSubtitleFile(ctx, &file.SubtitleFile)
		} else {
			err = s.Model.InsertEpisodeVideoFile(ctx, &file.VideoFile)
		}

		if err != nil {
			return fmt.Errorf("scanning media: %w", err)
		}
	}
}
