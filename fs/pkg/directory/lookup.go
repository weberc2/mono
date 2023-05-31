package directory

import (
	"fmt"

	. "github.com/weberc2/mono/fs/pkg/types"
)

func Lookup(fs *FileSystem, dirIno Ino, name string, out *FileInfo) error {
	var h Handle
	if err := Open(fs, dirIno, &h); err != nil {
		return fmt.Errorf("looking up `%s` in dir `%d`: %w", name, dirIno, err)
	}

	for {
		var info FileInfo
		if err := ReadNext(fs, &h, &info); err != nil {
			return fmt.Errorf(
				"looking up `%s` in dir `%d`: %w",
				name,
				dirIno,
				err,
			)
		}
		if info.Name == name {
			*out = info
			return nil
		}
	}
}
