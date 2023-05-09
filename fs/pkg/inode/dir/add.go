package dir

import (
	"fmt"

	"github.com/weberc2/mono/fs/pkg/encode"
	. "github.com/weberc2/mono/fs/pkg/types"
)

func AddEntry(
	fs *FileSystem,
	dir *Inode,
	entry *Inode,
	name []byte,
) error {
	if err := InsertEntry(
		fs,
		dir,
		entry,
		name,
		nil,
		encode.DirEntryHeaderSize+1, // skip the '.' entry
	); err != nil {
		return fmt.Errorf(
			"adding inode `%d` as entry in inode `%d` named `%s`: %w",
			entry.Ino,
			dir.Ino,
			name,
			err,
		)
	}
	return nil
}

func align4(x Byte) Byte {
	return (x + 0b11) &^ 0b11
}
