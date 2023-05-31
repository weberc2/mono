package filesystem

import (
	"fmt"
	"path/filepath"

	"github.com/weberc2/mono/fs/pkg/directory"
	. "github.com/weberc2/mono/fs/pkg/types"
)

type FileSystem = directory.FileSystem
type FileInfo = directory.FileInfo

func Lookup(fs *FileSystem, path string, out *FileInfo) error {
	if path[0] != '/' {
		return fmt.Errorf("looking up path `%s`: %w", path, NotAbsolutePathErr)
	}
	var chunks []string
	for {
		if path == "" {
			break
		}
		p, chunk := filepath.Split(path)
		path = p[:len(p)-1]
		if chunk != "" {
			chunks = append(chunks, chunk)
		}
	}

	ino := InoRoot
	for i := len(chunks) - 1; i >= 0; i-- {
		if err := directory.Lookup(fs, ino, chunks[i], out); err != nil {
			return fmt.Errorf("looking up path `%s`: %w", path, err)
		}
		ino = out.Ino
	}

	return nil
}

const (
	NotAbsolutePathErr ConstError = "not an absolute path"
)
