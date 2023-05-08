package dir

import (
	"fmt"

	"github.com/weberc2/mono/fs/pkg/inode/data"
	. "github.com/weberc2/mono/fs/pkg/types"
)

func RemoveEntry(writer *data.Writer, dir *Inode, freeSpace *FreeSpace) error {
	if err := WriteEntry(
		writer,
		dir,
		freeSpace.Offset,
		&DirEntry{
			Ino:      0,
			FileType: FileTypeInvalid,
			NameLen:  0,
		},
	); err != nil {
		return fmt.Errorf(
			"removing entry from inode `%d` at offset `%d` (prev offset "+
				"`%d`; next offset `%d`): %w",
			dir.Ino,
			freeSpace.Offset,
			freeSpace.PrevOffset,
			freeSpace.NextOffset,
			err,
		)
	}
	return nil
}
