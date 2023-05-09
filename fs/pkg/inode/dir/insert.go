package dir

import (
	"fmt"
	"log"
	"math"

	. "github.com/weberc2/mono/fs/pkg/types"
)

func InsertEntry(
	fs *FileSystem,
	dir *Inode,
	entry *Inode,
	name []byte,
	freeSpace *FreeSpace,
	lastOffset Byte,
) error {
	if len(name) > math.MaxUint8 {
		return fmt.Errorf(
			"inserting inode `%d` into dir `%d` with name `%s` at last "+
				"offset `%d`: %w",
			entry.Ino,
			dir.Ino,
			name,
			lastOffset,
			NameTooLongErr,
		)
	}

	if err := WriteEntry(
		fs.ReadWriter.Writer(),
		dir,
		freeSpace.Offset,
		&DirEntry{
			Ino:      entry.Ino,
			NameLen:  uint8(len(name)),
			FileType: entry.FileType,
			Name:     name,
		},
	); err != nil {
		return fmt.Errorf(
			"inserting inode `%d` into dir `%d` with name `%s` at last "+
				"offset `%d`: %w",
			entry.Ino,
			dir.Ino,
			name,
			lastOffset,
			err,
		)
	}
	log.Printf("writing entry `%d` into dir `%d` at offset `%d` with name `%s`", entry.Ino, dir.Ino, lastOffset, name)

	entry.LinksCount++
	if err := fs.InodeStore.Put(entry); err != nil {
		return fmt.Errorf(
			"inserting inode `%d` into dir `%d` with name `%s` at last "+
				"offset `%d`: %w",
			entry.Ino,
			dir.Ino,
			name,
			lastOffset,
			err,
		)
	}

	return nil
}
