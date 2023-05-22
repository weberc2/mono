package encode

import (
	stdmath "math"

	"github.com/weberc2/mono/fs/pkg/math"

	. "github.com/weberc2/mono/fs/pkg/types"
)

func DirEntrySize(nameLen uint8) Byte {
	return Byte(nameLen) + DirEntryHeaderSize
}

func DirEntryFreeSpace(entry *DirEntry) Byte {
	return Byte(entry.RecLen) - Byte(entry.NameLen) - DirEntryHeaderSize
}

func EncodeDirEntryHeader(entry *DirEntry, b *[DirEntryHeaderSize]byte) {
	p := b[:]
	nameLen := math.Min(len(entry.Name), stdmath.MaxUint8)
	putIno(p, dirEntryInoStart, entry.Ino)
	putU8(p, dirEntryFileTypeStart, uint8(entry.FileType))
	putU8(p, dirEntryNameLenStart, uint8(nameLen))
	putU16(p, DirEntryRecLenStart, entry.RecLen)
}

func DecodeDirEntryHeader(entry *DirEntry, b *[DirEntryHeaderSize]byte) {
	p := b[:]
	// NB: We are explicitly NOT validating the filetype here because it's
	// perfectly valid to have a zeroed-out DirEntry on disk (e.g., a direntry
	// gets deleted). Callers must validate if desired.
	entry.FileType = FileType(getU8(p, dirEntryFileTypeStart))
	entry.Ino = getIno(p, dirEntryInoStart)
	entry.NameLen = getU8(p, dirEntryNameLenStart)
	entry.RecLen = getU16(p, DirEntryRecLenStart)
}

func EncodeDirEntryRecLen(recLen uint16, b *[2]byte) {
	putU16(b[:], 0, recLen)
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

	DirEntryRecLenStart = dirEntryNameLenEnd
	dirEntryRecLenSize  = 2
	dirEntryRecLenEnd   = DirEntryRecLenStart + dirEntryRecLenSize

	DirEntryHeaderSize = dirEntryRecLenEnd
)
