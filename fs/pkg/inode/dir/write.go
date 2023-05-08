package dir

import (
	"fmt"

	"github.com/weberc2/mono/fs/pkg/encode"
	"github.com/weberc2/mono/fs/pkg/inode/data"
	. "github.com/weberc2/mono/fs/pkg/types"
)

func WriteEntry(
	writer *data.Writer,
	inode *Inode,
	offset Byte,
	entry *DirEntry,
) error {
	buf := make([]byte, encode.DirEntryHeaderSize+Byte(entry.NameLen))
	encode.EncodeDirEntryHeader(entry, (*[encode.DirEntryHeaderSize]byte)(buf))
	copy(buf[encode.DirEntryHeaderSize:], entry.Name)
	if _, err := writer.Write(inode, offset, buf); err != nil {
		return fmt.Errorf(
			"writing direntry `%s` (type `%s`) to inode `%d` at offset "+
				"`%d`: %w",
			entry.Name,
			entry.FileType,
			inode.Ino,
			offset,
			err,
		)
	}
	return nil
}
