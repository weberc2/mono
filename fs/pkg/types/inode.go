package types

import "fmt"

type Ino uint64

const (
	DirectBlocksCount Block = 12
	InodeSize         Byte  = 512
	InoSize           Byte  = 8
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
