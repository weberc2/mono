package subtitles

import (
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"testing"
	"testing/fstest"
)

func TestIterateShow(t *testing.T) {
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
				"shows/Test (2024)": &fstest.MapFile{Mode: fs.ModeDir},
			},
			directory: "shows",
			fileName:  "Test (2024)",
			wanted:    nil,
		},
		{
			name: "single-empty-season",
			fileSystem: fstest.MapFS{
				"shows/Test (2024)/Season 01": &fstest.MapFile{Mode: fs.ModeDir},
			},
			directory: "shows",
			fileName:  "Test (2024)",
			wanted:    nil,
		},
		{
			name: "multi-empty-seasons",
			fileSystem: fstest.MapFS{
				"shows/Test (2024)/Season 01": &fstest.MapFile{Mode: fs.ModeDir},
				"shows/Test (2024)/Season 02": &fstest.MapFile{Mode: fs.ModeDir},
			},
			directory: "shows",
			fileName:  "Test (2024)",
			wanted:    nil,
		},
		{
			name: "single-season-single-episode",
			fileSystem: fstest.MapFS{
				"shows/Test (2024)/Season 01/Episode 01.mkv": &emptyFile,
			},
			directory: "shows",
			fileName:  "Test (2024)",
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
			name: "single-season-single-episode-video-and-subtitles",
			fileSystem: fstest.MapFS{
				"shows/Test (2024)/Season 01/Episode 01.mkv":    &emptyFile,
				"shows/Test (2024)/Season 01/Episode 01.en.srt": &emptyFile,
			},
			directory: "shows",
			fileName:  "Test (2024)",
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
			name: "single-season-multi-episode",
			fileSystem: fstest.MapFS{
				"shows/Test (2024)/Season 01/Episode 01.mkv": &emptyFile,
				"shows/Test (2024)/Season 01/Episode 02.mkv": &emptyFile,
			},
			directory: "shows",
			fileName:  "Test (2024)",
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
		{
			name: "multi-season-multi-episode",
			fileSystem: fstest.MapFS{
				"shows/Test (2024)/Season 01/Episode 01.mkv": &emptyFile,
				"shows/Test (2024)/Season 01/Episode 02.mkv": &emptyFile,
				"shows/Test (2024)/Season 02/Episode 01.mkv": &emptyFile,
				"shows/Test (2024)/Season 02/Episode 02.mkv": &emptyFile,
			},
			directory: "shows",
			fileName:  "Test (2024)",
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
			}, {
				SubtitleFile: SubtitleFile[Episode]{
					VideoFile: VideoFile[Episode]{
						Filepath: "shows/Test (2024)/Season 02/Episode 01.mkv",
						ID: Episode{
							Title:   "Test",
							Year:    "2024",
							Season:  "02",
							Episode: "01",
						},
					},
				},
			}, {
				SubtitleFile: SubtitleFile[Episode]{VideoFile: VideoFile[Episode]{
					Filepath: "shows/Test (2024)/Season 02/Episode 02.mkv",
					ID: Episode{
						Title:   "Test",
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
			mediaFiles, err := collectShow(
				testCases[i].fileSystem,
				testCases[i].directory,
				testCases[i].fileName,
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

func collectShow(
	fsys fs.FS,
	dir string,
	fileName string,
) ([]MediaFile[Episode], error) {
	it, err := IterateShow(fsys, dir, fileName)
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
