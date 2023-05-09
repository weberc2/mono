package dir

import (
	"bytes"
	"fmt"

	"github.com/weberc2/mono/fs/pkg/encode"
	"github.com/weberc2/mono/fs/pkg/inode/data"
	. "github.com/weberc2/mono/fs/pkg/types"
)

func UnlinkInode(reader data.Reader, dir *Inode) (bool, error) {
	var offset Byte
	for offset < dir.Size {
		var entry DirEntry
		if err := ReadEntry(reader, dir, offset, &entry); err != nil {
			return false, fmt.Errorf(
				"unlinking directory inode `%d`: %w",
				dir.Ino,
				err,
			)
		}

		if entry.Ino == 0 &&
			!bytes.Equal(entry.Name, []byte(".")) &&
			!bytes.Equal(entry.Name, []byte("..")) {
			return true, nil
		}

		offset += encode.DirEntryHeaderSize + Byte(entry.NameLen)
	}

	return true, nil
}
