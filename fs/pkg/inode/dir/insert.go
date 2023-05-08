package dir

import (
	"fmt"
	"math"

	"github.com/weberc2/mono/fs/pkg/inode/data"
	. "github.com/weberc2/mono/fs/pkg/types"
)

func InsertEntry(
	writer *data.Writer,
	inodeStore InodeStore,
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
		writer,
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

	entry.LinksCount++
	if err := inodeStore.Put(entry); err != nil {
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
