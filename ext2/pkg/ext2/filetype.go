package ext2

import "fmt"

type FileType uint16

const (
	FileTypeRegular FileType = iota
	FileTypeDir
	FileTypeCharDev
	FileTypeBlockDev
	FileTypeFifo
	FileTypeSocket
	FileTypeSymlink

	InodeBufferSize = 128
)

func (fileType FileType) String() string {
	switch fileType {
	case FileTypeRegular:
		return "Regular"
	case FileTypeDir:
		return "Dir"
	case FileTypeCharDev:
		return "CharDev"
	case FileTypeBlockDev:
		return "BlockDev"
	case FileTypeFifo:
		return "Fifo"
	case FileTypeSocket:
		return "Socket"
	case FileTypeSymlink:
		return "Symlink"
	default:
		panic(fmt.Sprintf("unknown file type: %d", fileType))
	}
}

func (fileType FileType) Encode() uint16 {
	var tmp uint16
	switch fileType {
	case FileTypeFifo:
		tmp = 1
	case FileTypeCharDev:
		tmp = 2
	case FileTypeDir:
		tmp = 4
	case FileTypeBlockDev:
		tmp = 6
	case FileTypeRegular:
		tmp = 8
	case FileTypeSymlink:
		tmp = 10
	case FileTypeSocket:
		tmp = 12
	}
	return tmp << 12
}
