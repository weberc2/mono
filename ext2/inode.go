package main

type Inode struct {
	Ino        uint64
	Mode       Mode
	Attr       FileAttr
	Size       uint64
	Size512    uint32
	LinksCount uint16
	Flags      uint32
	Block      [15]uint32
	FileACL    uint32
}

type FileAttr struct {
	UID   uint32
	GID   uint32
	ATime uint32
	CTime uint32
	MTime uint32
	DTime uint32
}

type Mode struct {
	FileType     FileType
	SUID         bool
	SGID         bool
	Sticky       bool
	AccessRights uint16
}

type FileType uint

const (
	FileTypeRegular FileType = iota
	FileTypeDir
	FileTypeCharDev
	FileTypeBlockDev
	FileTypeFifo
	FileTypeSocket
	FileTypeSymlink
)
