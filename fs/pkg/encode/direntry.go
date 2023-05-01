package encode

import (
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

func DecodeDirEntryHeader(entry *DirEntry, b *[DirEntryHeaderSize]byte) {
	p := b[:]
	entry.Ino = getIno(p, dirEntryInoStart)
	entry.FileType = FileType(getU8(p, dirEntryFileTypeStart))
	entry.NameLen = getU8(p, dirEntryNameLenStart)
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
