package subtitles

import (
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"testing"
	"testing/fstest"
)

func TestIterateShows(t *testing.T) {
	testCases := []struct {
		name       string
		fileSystem fstest.MapFS
		directory  string
		wanted     []MediaFile[Episode]
		wantedErr  func(error) error
	}{
		{
			name: "empty",
			fileSystem: fstest.MapFS{
				"shows": &fstest.MapFile{Mode: fs.ModeDir},
			},
			directory: "shows",
			wanted:    nil,
		},
		{
			name: "single-empty-show",
			fileSystem: fstest.MapFS{
				"shows/Test (2024)": &fstest.MapFile{Mode: fs.ModeDir},
			},
			directory: "shows",
			wanted:    nil,
		},
		{
			name: "single-show-empty-season",
			fileSystem: fstest.MapFS{
				"shows/Test (2024)/Season 01": &fstest.MapFile{Mode: fs.ModeDir},
			},
			directory: "shows",
			wanted:    nil,
		},
		{
			name: "multi-empty-shows",
			fileSystem: fstest.MapFS{
				"shows/Test-01 (2024)": &fstest.MapFile{Mode: fs.ModeDir},
				"shows/Test-02 (2024)": &fstest.MapFile{Mode: fs.ModeDir},
			},
			directory: "shows",
			wanted:    nil,
		},
		{
			name: "single-show-single-season-single-episode",
			fileSystem: fstest.MapFS{
				"shows/Test (2024)/Season 01/Episode 01.mkv": &emptyFile,
			},
			directory: "shows",
			wanted: []MediaFile[Episode]{{
				SubtitleFile: SubtitleFile[Episode]{
					VideoFile: VideoFile[Episode]{
						Filepath: "shows/Test (2024)/Season 01/Episode 01.mkv",
						ID: Episode{
							Title:   "Test",
							Season:  "01",
							Episode: "01",
							Year:    "2024",
						},
					},
				},
			}},
		},
		{
			name: "multi-shows-multi-seasons-multi-episodes",
			fileSystem: fstest.MapFS{
				"shows/Test-01 (2024)/Season 01/Episode 01.mkv": &emptyFile,
				"shows/Test-01 (2024)/Season 01/Episode 02.mkv": &emptyFile,
				"shows/Test-01 (2024)/Season 02/Episode 01.mkv": &emptyFile,
				"shows/Test-01 (2024)/Season 02/Episode 02.mkv": &emptyFile,
				"shows/Test-02 (2024)/Season 01/Episode 01.mkv": &emptyFile,
				"shows/Test-02 (2024)/Season 01/Episode 02.mkv": &emptyFile,
				"shows/Test-02 (2024)/Season 02/Episode 01.mkv": &emptyFile,
				"shows/Test-02 (2024)/Season 02/Episode 02.mkv": &emptyFile,
			},
			directory: "shows",
			wanted: []MediaFile[Episode]{{
				SubtitleFile: SubtitleFile[Episode]{
					VideoFile: VideoFile[Episode]{
						Filepath: "shows/Test-01 (2024)/Season 01/Episode 01.mkv",
						ID: Episode{
							Title:   "Test-01",
							Year:    "2024",
							Season:  "01",
							Episode: "01",
						},
					},
				},
			}, {
				SubtitleFile: SubtitleFile[Episode]{
					VideoFile: VideoFile[Episode]{
						Filepath: "shows/Test-01 (2024)/Season 01/Episode 02.mkv",
						ID: Episode{
							Title:   "Test-01",
							Year:    "2024",
							Season:  "01",
							Episode: "02",
						},
					},
				},
			}, {
				SubtitleFile: SubtitleFile[Episode]{
					VideoFile: VideoFile[Episode]{
						Filepath: "shows/Test-01 (2024)/Season 02/Episode 01.mkv",
						ID: Episode{
							Title:   "Test-01",
							Year:    "2024",
							Season:  "02",
							Episode: "01",
						},
					},
				},
			}, {
				SubtitleFile: SubtitleFile[Episode]{
					VideoFile: VideoFile[Episode]{
						Filepath: "shows/Test-01 (2024)/Season 02/Episode 02.mkv",
						ID: Episode{
							Title:   "Test-01",
							Year:    "2024",
							Season:  "02",
							Episode: "02",
						},
					},
				},
			}, {
				SubtitleFile: SubtitleFile[Episode]{
					VideoFile: VideoFile[Episode]{
						Filepath: "shows/Test-02 (2024)/Season 01/Episode 01.mkv",
						ID: Episode{
							Title:   "Test-02",
							Year:    "2024",
							Season:  "01",
							Episode: "01",
						},
					},
				},
			}, {
				SubtitleFile: SubtitleFile[Episode]{
					VideoFile: VideoFile[Episode]{
						Filepath: "shows/Test-02 (2024)/Season 01/Episode 02.mkv",
						ID: Episode{
							Title:   "Test-02",
							Year:    "2024",
							Season:  "01",
							Episode: "02",
						},
					},
				},
			}, {
				SubtitleFile: SubtitleFile[Episode]{
					VideoFile: VideoFile[Episode]{
						Filepath: "shows/Test-02 (2024)/Season 02/Episode 01.mkv",
						ID: Episode{
							Title:   "Test-02",
							Year:    "2024",
							Season:  "02",
							Episode: "01",
						},
					},
				},
			}, {
				SubtitleFile: SubtitleFile[Episode]{
					VideoFile: VideoFile[Episode]{
						Filepath: "shows/Test-02 (2024)/Season 02/Episode 02.mkv",
						ID: Episode{
							Title:   "Test-02",
							Year:    "2024",
							Season:  "02",
							Episode: "02",
						},
					},
				},
			}},
		},
	}

	for i := range testCases {
		t.Run(testCases[i].name, func(t *testing.T) {
			mediaFiles, err := collectShows(
				testCases[i].fileSystem,
				testCases[i].directory,
			)
			if testCases[i].wantedErr == nil {
				if err != nil {
					t.Fatalf("unexpected error parsing show media file: %v", err)
				}
				return
			} else if err := testCases[i].wantedErr(err); err != nil {
				t.Fatal(err)
			}

			if len(mediaFiles) != len(testCases[i].wanted) {
				goto ERROR
			}

			for j := range mediaFiles {
				if mediaFiles[j] != testCases[i].wanted[j] {
					goto ERROR
				}
			}

			return

		ERROR:
			wanted, err := json.Marshal(&testCases[i].wanted)
			if err != nil {
				t.Fatalf("unexpected error marshaling `wanted`: %v", err)
			}
			found, err := json.Marshal(&mediaFiles)
			if err != nil {
				t.Fatalf("unexpected error marshaling `found`: %v", err)
			}

			t.Fatalf(
				"comparing media files: wanted `%s`; found `%s`",
				wanted,
				found,
			)
		})
	}
}

func collectShows(fsys fs.FS, dir string) ([]MediaFile[Episode], error) {
	it, err := IterateShows(fsys, dir)
	if err != nil {
		return nil, err
	}

	var out []MediaFile[Episode]
	for {
		mf, err := it.Next(fsys)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		out = append(out, mf)
	}

	return out, nil
}
