package subtitles

import (
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"sort"
	"testing"
	"testing/fstest"
)

func TestIterateSeason(t *testing.T) {
	testCases := []struct {
		name       string
		fileSystem fstest.MapFS
		directory  string
		fileName   string
		wanted     []MediaFile[Episode]
		wantedErr  func(error) error
	}{
		{
			name: "empty",
			fileSystem: fstest.MapFS{
				"shows/Test (2024)/Season 01": &fstest.MapFile{
					Mode: fs.ModeDir,
				},
			},
			directory: "shows/Test (2024)",
			fileName:  "Season 01",
			wanted:    nil,
		},
		{
			name: "single-video-file",
			fileSystem: fstest.MapFS{
				"shows/Test (2024)/Season 01/Episode 01.mkv": &emptyFile,
			},
			directory: "shows/Test (2024)",
			fileName:  "Season 01",
			wanted: []MediaFile[Episode]{{
				SubtitleFile: SubtitleFile[Episode]{
					VideoFile: VideoFile[Episode]{
						Filepath: "shows/Test (2024)/Season 01/Episode 01.mkv",
						ID: Episode{
							Title:   "Test",
							Year:    "2024",
							Season:  "01",
							Episode: "01",
						},
					},
				},
			}},
		},
		{
			name: "single-subtitle-file",
			fileSystem: fstest.MapFS{
				"shows/Test (2024)/Season 01/Episode 01.en.srt": &emptyFile,
			},
			directory: "shows/Test (2024)",
			fileName:  "Season 01",
			wanted: []MediaFile[Episode]{{
				SubtitleFile: SubtitleFile[Episode]{
					VideoFile: VideoFile[Episode]{
						Filepath: "shows/Test (2024)/Season 01/Episode 01.en.srt",
						ID: Episode{
							Title:   "Test",
							Year:    "2024",
							Season:  "01",
							Episode: "01",
						},
					},
					Language: "en",
				},
				IsSubtitle: true,
			}},
		},
		{
			name: "video-and-subtitle",
			fileSystem: fstest.MapFS{
				"shows/Test (2024)/Season 01/Episode 01.en.srt": &emptyFile,
				"shows/Test (2024)/Season 01/Episode 01.mkv":    &emptyFile,
			},
			directory: "shows/Test (2024)",
			fileName:  "Season 01",
			wanted: []MediaFile[Episode]{{
				SubtitleFile: SubtitleFile[Episode]{
					VideoFile: VideoFile[Episode]{
						Filepath: "shows/Test (2024)/Season 01/Episode 01.en.srt",
						ID: Episode{
							Title:   "Test",
							Year:    "2024",
							Season:  "01",
							Episode: "01",
						},
					},
					Language: "en",
				},
				IsSubtitle: true,
			}, {
				SubtitleFile: SubtitleFile[Episode]{
					VideoFile: VideoFile[Episode]{
						Filepath: "shows/Test (2024)/Season 01/Episode 01.mkv",
						ID: Episode{
							Title:   "Test",
							Year:    "2024",
							Season:  "01",
							Episode: "01",
						},
					},
				},
			}},
		},
		{
			name: "multi-episodes",
			fileSystem: fstest.MapFS{
				"shows/Test (2024)/Season 01/Episode 01.mkv": &emptyFile,
				"shows/Test (2024)/Season 01/Episode 02.mkv": &emptyFile,
			},
			directory: "shows/Test (2024)",
			fileName:  "Season 01",
			wanted: []MediaFile[Episode]{{
				SubtitleFile: SubtitleFile[Episode]{
					VideoFile: VideoFile[Episode]{
						Filepath: "shows/Test (2024)/Season 01/Episode 01.mkv",
						ID: Episode{
							Title:   "Test",
							Year:    "2024",
							Season:  "01",
							Episode: "01",
						},
					},
				},
			}, {
				SubtitleFile: SubtitleFile[Episode]{
					VideoFile: VideoFile[Episode]{
						Filepath: "shows/Test (2024)/Season 01/Episode 02.mkv",
						ID: Episode{
							Title:   "Test",
							Year:    "2024",
							Season:  "01",
							Episode: "02",
						},
					},
				},
			}},
		},
	}

	for i := range testCases {
		t.Run(testCases[i].name, func(t *testing.T) {
			mediaFiles, err := collectSeason(
				testCases[i].fileSystem,
				testCases[i].directory,
				testCases[i].fileName,
			)
			if testCases[i].wantedErr == nil {
				if err != nil {
					t.Fatalf("unexpected error parsing show media file: %v", err)
				}
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

func collectSeason(fsys fs.FS, dir, fileName string) ([]MediaFile[Episode], error) {
	it, err := IterateSeason(fsys, "Test", "2024", dir, fileName)
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

	sort.Slice(out, func(i, j int) bool {
		return out[i].Filepath < out[j].Filepath
	})
	return out, nil
}
