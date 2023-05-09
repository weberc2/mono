package dir

import (
	"fmt"

	. "github.com/weberc2/mono/fs/pkg/types"
)

func Open(fs *FileSystem, ino Ino, handle *Handle) error {
	var inode Inode
	if err := fs.InodeStore.Get(ino, &inode); err != nil {
		return fmt.Errorf("opening inode `%d` as directory: %w", ino, err)
	}

	if inode.FileType != FileTypeDir {
		return fmt.Errorf(
			"opening inode `%d` as directory: %w",
			ino,
			NotADirErr,
		)
	}

	handle.ino = ino
	return nil
}
