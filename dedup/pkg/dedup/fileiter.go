package dedup

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
)

type FileIter struct {
	directory   string
	directories []string
	entries     []fs.DirEntry
	cursor      int
}

func NewFileIter(directory string) (iter FileIter) {
	iter.directories = []string{directory}
	return
}

func (iter *FileIter) Next() (result Option[Result[File]]) {
	for {
		// loop over the remaining entries until we hit a file. if the
		// entries point to a directory, push it onto the queue
		for iter.cursor < len(iter.entries) {
			path := filepath.Join(
				iter.directory,
				iter.entries[iter.cursor].Name(),
			)

			if iter.entries[iter.cursor].IsDir() {
				iter.directories = append(iter.directories, path)
				iter.cursor++
				continue
			}

			var info fs.FileInfo
			info, result.Some.Err = iter.entries[iter.cursor].Info()
			if result.Some.Err != nil {
				result.Some.Err = fmt.Errorf(
					"fetching info for file `%s`: %w",
					path,
					result.Some.Err,
				)
				result.Exists = true
				return
			}

			result.Some.OK.Path = path
			result.Some.OK.Ino = info.Sys().(*syscall.Stat_t).Ino
			result.Some.OK.Size = info.Size()
			result.Exists = true
			iter.cursor++
			return
		}

		// LOAD THE NEXT DIRECTORY, IF ANY
		iter.cursor = 0

		// check to see if there are more directories to read, otherwise EOF
		if len(iter.directories) < 1 {
			return
		}

		// pop off the next directory
		iter.directory = iter.directories[0]
		iter.directories = iter.directories[1:]

		// read the next directory
		if iter.entries, result.Some.Err = os.ReadDir(
			iter.directory,
		); result.Some.Err != nil {
			result.Some.Err = fmt.Errorf(
				"reading dir `%s`: %w",
				iter.directory,
				result.Some.Err,
			)
			result.Exists = true
			return
		}
	}
}
