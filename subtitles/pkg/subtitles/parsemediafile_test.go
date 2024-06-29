package subtitles

import (
	"encoding/json"
	"testing"
	"testing/fstest"
)

func TestParseShowMediaFile(t *testing.T) {
	testCases := []struct {
		name       string
		fileSystem fstest.MapFS
		directory  string
		fileName   string
		wanted     MediaFile
		wantedErr  func(error) error
	}{
		{
			name: "mkv-simple",
			fileSystem: fstest.MapFS{
				"shows/Test (2024)/Season 01/Test S01E01.mkv": &empty64kFile,
			},
			directory: "shows/Test (2024)/Season 01/",
			fileName:  "Test S01E01.mkv",
			wanted: MediaFile{
				Filepath:  "shows/Test (2024)/Season 01/Test S01E01.mkv",
				Type:      MediaTypeShow,
				Kind:      MediaFileKindVideo,
				Title:     "Test",
				Season:    "01",
				Episode:   "01",
				Mediahash: empty64kFileMediahash,
			},
		},
		{
			name: "srt-simple",
			fileSystem: fstest.MapFS{
				"shows/Test (2024)/Season 01/Test S01E01.en.srt": &empty64kFile,
			},
			directory: "shows/Test (2024)/Season 01/",
			fileName:  "Test S01E01.en.srt",
			wanted: MediaFile{
				Filepath:         "shows/Test (2024)/Season 01/Test S01E01.en.srt",
				Type:             MediaTypeShow,
				Kind:             MediaFileKindSubtitle,
				Title:            "Test",
				Season:           "01",
				Episode:          "01",
				SubtitleLanguage: "en",
			},
		},
		{
			name: "srt-lang-locale",
			fileSystem: fstest.MapFS{
				"shows/Test (2024)/Season 01/Test S01E01.en-US.srt": &empty64kFile,
			},
			directory: "shows/Test (2024)/Season 01/",
			fileName:  "Test S01E01.en-US.srt",
			wanted: MediaFile{
				Filepath:         "shows/Test (2024)/Season 01/Test S01E01.en-US.srt",
				Type:             MediaTypeShow,
				Kind:             MediaFileKindSubtitle,
				Title:            "Test",
				Season:           "01",
				Episode:          "01",
				SubtitleLanguage: "en-US",
			},
		},
		{
			name: "the-americans",
			fileSystem: fstest.MapFS{
				"The Americans S01E01.mkv": &empty64kFile,
			},
			fileName: "The Americans S01E01.mkv",
			wanted: MediaFile{
				Filepath:  "The Americans S01E01.mkv",
				Type:      MediaTypeShow,
				Kind:      MediaFileKindVideo,
				Title:     "The Americans",
				Season:    "01",
				Episode:   "01",
				Mediahash: empty64kFileMediahash,
			},
		},
	}
	for i := range testCases {
		t.Run(testCases[i].name, func(t *testing.T) {
			mediaFile, err := ParseShowMediaFile(
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

			if mediaFile != testCases[i].wanted {
				wanted, err := json.Marshal(&testCases[i].wanted)
				if err != nil {
					t.Fatalf("unexpected error marshaling `wanted`: %v", err)
				}
				found, err := json.Marshal(&mediaFile)
				if err != nil {
					t.Fatalf("unexpected error marshaling `found`: %v", err)
				}

				t.Fatalf(
					"comparing media files: wanted `%s`; found `%s`",
					wanted,
					found,
				)
			}
		})
	}
}

var empty64kFile = fstest.MapFile{
	Data: make([]byte, ChunkSize),
}
var empty64kFileMediahash = func() string {
	mapfs := fstest.MapFS{"foo": &empty64kFile}
	file, err := mapfs.Open("foo")
	if err != nil {
		panic(err)
	}
	hash, err := Mediahash(file)
	if err != nil {
		panic(err)
	}
	return hash
}()
