package types

import (
	"fmt"
)

type Ino uint64

const (
	DirectBlocksCount Block = 12
	InodeSize         Byte  = 512
	InoSize           Byte  = 8
	InoNil            Ino   = 0
	InoRoot           Ino   = 1
)

type Inode struct {
	Ino                 Ino
	FileType            FileType
	Size                Byte
	LinksCount          uint16
	DirectBlocks        [DirectBlocksCount]Block
	SinglyIndirectBlock Block
	DoublyIndirectBlock Block
	TriplyIndirectBlock Block
}

type FileType uint8

const (
	FileTypeInvalid FileType = iota
	FileTypeRegular
	FileTypeDir
	FileTypeCharDev
	FileTypeBlockDev
	FileTypeFifo
	FileTypeSocket
	FileTypeSymlink
)

func (ft FileType) String() string {
	switch ft {
	case FileTypeInvalid:
		return "Invalid"
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
		panic(fmt.Sprintf("invalid file type: `%d`", ft))
	}
}

func (ft FileType) MarshalJSON() ([]byte, error) {
	s := ft.String()
	out := make([]byte, len(s)+2)
	out[0] = '"'
	out[len(out)-1] = '"'
	copy(out[1:], s)
	return out, nil
}

func (ft FileType) Validate() error {
	if ft <= FileTypeInvalid || ft > FileTypeSymlink {
		return fmt.Errorf(
			"validating file type `%d`: %w",
			ft,
			InvalidFileTypeErr,
		)
	}
	return nil
}

const (
	InvalidFileTypeErr ConstError = "invalid file type"
)
