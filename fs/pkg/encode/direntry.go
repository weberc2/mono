package encode

import (
	"fmt"
	stdmath "math"

	"github.com/weberc2/mono/fs/pkg/math"

	. "github.com/weberc2/mono/fs/pkg/types"
)

func EncodeDirEntryHeader(entry *DirEntry, b *[DirEntryHeaderSize]byte) {
	p := b[:]
	nameLen := math.Min(len(entry.Name), stdmath.MaxUint8)
	putIno(p, dirEntryInoStart, entry.Ino)
	putU8(p, dirEntryFileTypeStart, uint8(entry.FileType))
	putU8(p, dirEntryNameLenStart, uint8(nameLen))
}

func DecodeDirEntryHeader(entry *DirEntry, b *[DirEntryHeaderSize]byte) error {
	p := b[:]
	ft := FileType(getU8(p, dirEntryFileTypeStart))
	if err := ft.Validate(); err != nil {
		return fmt.Errorf("decoding direntry header: %w", err)
	}
	entry.FileType = ft
	entry.Ino = getIno(p, dirEntryInoStart)
	entry.NameLen = getU8(p, dirEntryNameLenStart)
	return nil
}

const (
	dirEntryInoStart = 0
	dirEntryInoSize  = InoSize
	dirEntryInoEnd   = dirEntryInoStart + dirEntryInoSize

	dirEntryFileTypeStart = dirEntryInoEnd
	dirEntryFileTypeSize  = 1
	dirEntryFileTypeEnd   = dirEntryFileTypeStart + dirEntryFileTypeSize

	dirEntryNameLenStart = dirEntryFileTypeEnd
	dirEntryNameLenSize  = 1
	dirEntryNameLenEnd   = dirEntryNameLenStart + dirEntryNameLenSize

	DirEntryHeaderSize = dirEntryNameLenEnd
)
