package subtitles

import (
	"testing"
	"testing/fstest"
)

func TestParseShowMediaFile(t *testing.T) {
	testCases := []struct {
		name           string
		fileSystem     fstest.MapFS
		directory      string
		fileName       string
		wantedEpisode  string
		wantedLanguage string
		wantedSubtitle bool
		wantedErr      func(error) error
	}{
		{
			name: "mkv-simple",
			fileSystem: fstest.MapFS{
				"shows/Test (2024)/Season 01/Test S01E01.mkv": &emptyFile,
			},
			directory:     "shows/Test (2024)/Season 01/",
			fileName:      "Episode 01.mkv",
			wantedEpisode: "01",
		},
		{
			name: "srt-simple",
			fileSystem: fstest.MapFS{
				"shows/Test (2024)/Season 01/Test S01E01.en.srt": &emptyFile,
			},
			directory:      "shows/Test (2024)/Season 01/",
			fileName:       "Episode 01.en.srt",
			wantedEpisode:  "01",
			wantedLanguage: "en",
			wantedSubtitle: true,
		},
		{
			name: "the-americans",
			fileSystem: fstest.MapFS{
				"The Americans S01E01.mkv": &emptyFile,
			},
			fileName:      "Episode 01.mkv",
			wantedEpisode: "01",
		},
	}
	for i := range testCases {
		t.Run(testCases[i].name, func(t *testing.T) {
			epi, lng, sub, err := ParseShowMediaFile(
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

			if epi != testCases[i].wantedEpisode {
				t.Fatalf(
					"episode: wanted `%s`; found `%s`",
					testCases[i].wantedEpisode,
					epi,
				)
			}

			if lng != testCases[i].wantedLanguage {
				t.Fatalf(
					"language: wanted `%s`; found `%s`",
					testCases[i].wantedLanguage,
					lng,
				)
			}

			if sub != testCases[i].wantedSubtitle {
				t.Fatalf(
					"subtitle: wanted `%t`; found `%t`",
					testCases[i].wantedSubtitle,
					sub,
				)
			}
		})
	}
}

var emptyFile = fstest.MapFile{}
