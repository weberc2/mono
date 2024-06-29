package subtitles

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
)

type ShowsIterator struct {
	Directory   string
	Files       []fs.DirEntry
	CurrentFile int
	CurrentShow ShowIterator
}

func IterateShows(fsys fs.FS, directory string) (it ShowsIterator, err error) {
	it.Directory = directory
	if it.Files, err = fs.ReadDir(fsys, directory); err != nil {
		err = fmt.Errorf("parsing shows: %w", err)
	}
	return
}

func (iter *ShowsIterator) Next(fsys fs.FS) (mf MediaFile, err error) {
TOP:
	if mf, err = iter.CurrentShow.Next(fsys); err == nil {
		return
	}

	if !errors.Is(err, io.EOF) {
		err = fmt.Errorf("parsing shows: %w", err)
		return
	}

NEXT_FILE:
	if iter.CurrentFile < len(iter.Files) {
		file := iter.Files[iter.CurrentFile]
		if !file.IsDir() {
			slog.Warn(
				"parsing shows: skipping unexpected non-directory",
				"showsDirectory", iter.Directory,
				"fileName", file.Name(),
			)
			iter.CurrentFile++
			goto NEXT_FILE
		}

		if iter.CurrentShow, err = IterateShow(
			fsys,
			iter.Directory,
			file.Name(),
		); err != nil {
			err = fmt.Errorf("parsing shows: %w", err)
			return
		}

		iter.CurrentFile++
		goto TOP
	}

	// if we've iterated through all of the shows, return `io.EOF`
	err = io.EOF
	return
}
