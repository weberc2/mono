package subtitles

import (
	"encoding/json"
	"errors"
	"fmt"
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
		wanted     []MediaFile
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
				"shows/Test (2024)/Season 01/Test S01E01.mkv": &empty64kFile,
			},
			directory: "shows",
			fileName:  "Test (2024)",
			wanted: []MediaFile{{
				Filepath:  "shows/Test (2024)/Season 01/Test S01E01.mkv",
				Type:      MediaTypeShow,
				Kind:      MediaFileKindVideo,
				Title:     "Test",
				Season:    "01",
				Episode:   "01",
				Mediahash: empty64kFileMediahash,
			}},
		},
		{
			name: "single-season-single-episode-video-and-subtitles",
			fileSystem: fstest.MapFS{
				"shows/Test (2024)/Season 01/Test S01E01.mkv":    &empty64kFile,
				"shows/Test (2024)/Season 01/Test S01E01.en.srt": &empty64kFile,
			},
			directory: "shows",
			fileName:  "Test (2024)",
			wanted: []MediaFile{{
				Filepath:  "shows/Test (2024)/Season 01/Test S01E01.mkv",
				Type:      MediaTypeShow,
				Kind:      MediaFileKindVideo,
				Title:     "Test",
				Season:    "01",
				Episode:   "01",
				Mediahash: empty64kFileMediahash,
			}, {
				Filepath:         "shows/Test (2024)/Season 01/Test S01E01.en.srt",
				Type:             MediaTypeShow,
				Kind:             MediaFileKindSubtitle,
				Title:            "Test",
				Season:           "01",
				Episode:          "01",
				SubtitleLanguage: "en",
			}},
		},
		{
			name: "single-season-multi-episode",
			fileSystem: fstest.MapFS{
				"shows/Test (2024)/Season 01/Test S01E01.mkv": &empty64kFile,
				"shows/Test (2024)/Season 01/Test S01E02.mkv": &empty64kFile,
			},
			directory: "shows",
			fileName:  "Test (2024)",
			wanted: []MediaFile{{
				Filepath:  "shows/Test (2024)/Season 01/Test S01E01.mkv",
				Type:      MediaTypeShow,
				Kind:      MediaFileKindVideo,
				Title:     "Test",
				Season:    "01",
				Episode:   "01",
				Mediahash: empty64kFileMediahash,
			}, {
				Filepath:  "shows/Test (2024)/Season 01/Test S01E02.mkv",
				Type:      MediaTypeShow,
				Kind:      MediaFileKindVideo,
				Title:     "Test",
				Season:    "01",
				Episode:   "02",
				Mediahash: empty64kFileMediahash,
			}},
		},
		{
			name: "multi-season-multi-episode",
			fileSystem: fstest.MapFS{
				"shows/Test (2024)/Season 01/Test S01E01.mkv": &empty64kFile,
				"shows/Test (2024)/Season 01/Test S01E02.mkv": &empty64kFile,
				"shows/Test (2024)/Season 02/Test S02E01.mkv": &empty64kFile,
				"shows/Test (2024)/Season 02/Test S02E02.mkv": &empty64kFile,
			},
			directory: "shows",
			fileName:  "Test (2024)",
			wanted: []MediaFile{{
				Filepath:  "shows/Test (2024)/Season 01/Test S01E01.mkv",
				Type:      MediaTypeShow,
				Kind:      MediaFileKindVideo,
				Title:     "Test",
				Season:    "01",
				Episode:   "01",
				Mediahash: empty64kFileMediahash,
			}, {
				Filepath:  "shows/Test (2024)/Season 01/Test S01E02.mkv",
				Type:      MediaTypeShow,
				Kind:      MediaFileKindVideo,
				Title:     "Test",
				Season:    "01",
				Episode:   "02",
				Mediahash: empty64kFileMediahash,
			}, {
				Filepath:  "shows/Test (2024)/Season 02/Test S01E01.mkv",
				Type:      MediaTypeShow,
				Kind:      MediaFileKindVideo,
				Title:     "Test",
				Season:    "02",
				Episode:   "01",
				Mediahash: empty64kFileMediahash,
			}, {
				Filepath:  "shows/Test (2024)/Season 02/Test S01E02.mkv",
				Type:      MediaTypeShow,
				Kind:      MediaFileKindVideo,
				Title:     "Test",
				Season:    "02",
				Episode:   "02",
				Mediahash: empty64kFileMediahash,
			}},
		},
		{
			name: "mismatched-season-number",
			fileSystem: fstest.MapFS{
				"shows/Test (2024)/Season 01/Test S02E01.mkv": &empty64kFile,
			},
			directory: "shows",
			fileName:  "Test (2024)",
			wanted:    nil,
			wantedErr: func(err error) error {
				wanted := IncorrectSeasonNumberErr{
					WantedSeason:    "01",
					MediaFileSeason: "02",
					MediaFileName:   "Test S02E01.mkv",
				}
				var e *IncorrectSeasonNumberErr
				if errors.As(err, &e) {
					if *e == wanted {
						return nil
					}
				}
				return fmt.Errorf("wanted error `%v`; found `%v`", &wanted, err)
			},
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

func collectShow(fsys fs.FS, dir, fileName string) ([]MediaFile, error) {
	it, err := IterateShow(fsys, dir, fileName)
	if err != nil {
		return nil, err
	}

	var out []MediaFile
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
