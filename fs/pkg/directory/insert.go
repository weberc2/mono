package directory

import (
	"fmt"
	"math"

	"github.com/weberc2/mono/fs/pkg/encode"
	. "github.com/weberc2/mono/fs/pkg/types"
)

func InsertEntry(
	fs *FileSystem,
	dir *Inode,
	entry *Inode,
	name string,
	offset Byte,
) error {
	if len(name) > math.MaxUint8 {
		return fmt.Errorf(
			"inserting inode `%d` into dir `%d` with name `%s` at offset "+
				"`%d`: %w",
			entry.Ino,
			dir.Ino,
			name,
			offset,
			NameTooLongErr,
		)
	}

	if err := WriteEntry(
		fs.ReadWriter.Writer(),
		dir,
		offset,
		&DirEntry{
			Ino:      entry.Ino,
			NameLen:  uint8(len(name)),
			RecLen:   uint16(encode.DirEntryHeaderSize) + uint16(len(name)),
			FileType: entry.FileType,
			Name:     name,
		},
	); err != nil {
		return fmt.Errorf(
			"inserting inode `%d` into dir `%d` with name `%s` at offset "+
				"`%d`: %w",
			entry.Ino,
			dir.Ino,
			name,
			offset,
			err,
		)
	}

	entry.LinksCount++
	if err := fs.InodeStore.Put(entry); err != nil {
		return fmt.Errorf(
			"inserting inode `%d` into dir `%d` with name `%s` at offset "+
				"`%d`: %w",
			entry.Ino,
			dir.Ino,
			name,
			offset,
			err,
		)
	}

	return nil
}
