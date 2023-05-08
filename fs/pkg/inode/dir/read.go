package dir

import (
	"fmt"

	"github.com/weberc2/mono/fs/pkg/encode"
	"github.com/weberc2/mono/fs/pkg/inode/data"
	. "github.com/weberc2/mono/fs/pkg/types"
)

func ReadEntry(
	reader *data.Reader,
	inode *Inode,
	offset Byte,
	out *DirEntry,
) error {
	buf := new([encode.DirEntryHeaderSize]byte)
	if _, err := reader.Read(
		inode,
		offset,
		buf[:],
	); err != nil {
		return fmt.Errorf(
			"reading direntry for inode `%d` at offset `%d`: %w",
			inode.Ino,
			offset,
			err,
		)
	}

	if err := encode.DecodeDirEntryHeader(out, buf); err != nil {
		return fmt.Errorf(
			"reading direntry for inode `%d` at offset `%d`: %w",
			inode.Ino,
			offset,
			err,
		)
	}

	out.Name = make([]byte, out.NameLen)
	if _, err := reader.Read(
		inode,
		offset+encode.DirEntryHeaderSize,
		out.Name,
	); err != nil {
		return fmt.Errorf(
			"reading direntry for inode `%d` at offset `%d`: %w",
			inode.Ino,
			offset,
			err,
		)
	}

	return nil
}
