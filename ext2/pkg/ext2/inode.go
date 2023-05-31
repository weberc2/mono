package ext2

import "fmt"

type Inode struct {
	Ino        Ino
	Mode       Mode
	Attr       FileAttr
	Size       uint64
	Size512    uint32
	LinksCount uint16
	Flags      uint32
	Block      [15]uint32
	FileACL    uint32
}

func DecodeInode(ino Ino, revLevel RevLevel, b *[InodeBufferSize]byte) (Inode, error) {
	mode, err := DecodeInodeMode(DecodeUint16(b[0], b[1]))
	if err != nil {
		return Inode{}, fmt.Errorf("decoding inode `%#x`: %w", ino, err)
	}

	sizeLow := uint64(DecodeUint32(b[4], b[5], b[6], b[7]))
	sizeHigh := uint64(0)
	if revLevel > RevLevelStatic && mode.FileType == FileTypeRegular {
		sizeHigh = uint64(DecodeUint32(b[108], b[109], b[110], b[111]))
	}

	uidLow := uint32(DecodeUint16(b[2], b[3]))
	uidHigh := uint32(DecodeUint16(b[120], b[121]))
	gidLow := uint32(DecodeUint16(b[24], b[25]))
	gidHigh := uint32(DecodeUint16(b[122], b[123]))

	var block [15]uint32
	for i := range block {
		base := 40 + 4*i
		block[i] = DecodeUint32(b[base], b[base+1], b[base+2], b[base+3])
	}

	return Inode{
		Ino:  ino,
		Mode: mode,
		Attr: FileAttr{
			UID:   uidLow + (uidHigh << 16),
			GID:   gidLow + (gidHigh << 16),
			ATime: DecodeUint32(b[8], b[9], b[10], b[11]),
			CTime: DecodeUint32(b[12], b[13], b[14], b[15]),
			MTime: DecodeUint32(b[16], b[17], b[18], b[19]),
			DTime: DecodeUint32(b[20], b[21], b[22], b[23]),
		},
		Size:       sizeLow + (sizeHigh << 32),
		Size512:    DecodeUint32(b[28], b[29], b[30], b[31]),
		LinksCount: DecodeUint16(b[26], b[27]),
		Flags:      DecodeUint32(b[32], b[33], b[34], b[35]),
		Block:      block,
		FileACL:    DecodeUint32(b[104], b[105], b[106], b[107]),
	}, nil
}

type ErrFileSizeTooLargeForStaticRevLevel struct {
	FileSize uint64
}

func (err ErrFileSizeTooLargeForStaticRevLevel) Error() string {
	return fmt.Sprintf(
		"file size cannot exceed 32 bits for rev level %d; found file "+
			"size `%#x`",
		RevLevelStatic,
		err.FileSize,
	)
}

func (inode *Inode) Encode(revLevel RevLevel, b *[InodeBufferSize]byte) error {
	EncodeUint16(inode.Mode.Encode(), b[0:])

	EncodeUint16(uint16(inode.Attr.UID&0xffff), b[2:])
	EncodeUint16(uint16((inode.Attr.UID>>16)&0xffff), b[120:])
	EncodeUint16(uint16(inode.Attr.GID&0xffff), b[24:])
	EncodeUint16(uint16((inode.Attr.GID>>16)&0xffff), b[122:])

	EncodeUint32(uint32(inode.Size&0xffffffff), b[4:])
	if (inode.Size>>32) != 0 && revLevel == RevLevelStatic {
		return fmt.Errorf(
			"encoding inode `%#x`: %w",
			inode.Ino,
			ErrFileSizeTooLargeForStaticRevLevel{inode.Size},
		)
	} else {
		EncodeUint32(uint32((inode.Size>>32)&0xffffffff), b[108:])
	}

	for i := range inode.Block {
		EncodeUint32(inode.Block[i], b[40+4*i:])
	}
	EncodeUint32(inode.Attr.ATime, b[8:])
	EncodeUint32(inode.Attr.CTime, b[12:])
	EncodeUint32(inode.Attr.MTime, b[16:])
	EncodeUint32(inode.Attr.DTime, b[20:])
	EncodeUint16(inode.LinksCount, b[26:])
	EncodeUint32(inode.Size512, b[28:])
	EncodeUint32(inode.Flags, b[32:])
	EncodeUint32(inode.FileACL, b[104:])

	return nil
}

type Ino uint64

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

type ErrUnknownFileType struct {
	FoundNibble uint16
}

func (err *ErrUnknownFileType) Error() string {
	return fmt.Sprintf("unknown file type nibble: %d", err.FoundNibble)
}

func DecodeInodeMode(mode uint16) (Mode, error) {
	typeNibble := (mode & 0xf000) >> 12
	var fileType FileType
	switch typeNibble {
	case 1:
		fileType = FileTypeFifo
	case 2:
		fileType = FileTypeCharDev
	case 4:
		fileType = FileTypeDir
	case 6:
		fileType = FileTypeBlockDev
	case 8:
		fileType = FileTypeRegular
	case 10:
		fileType = FileTypeSymlink
	case 12:
		fileType = FileTypeSocket
	default:
		return Mode{}, fmt.Errorf(
			"decoding inode mode `%d`: %w",
			mode,
			ErrUnknownFileType{typeNibble},
		)
	}

	return Mode{
		FileType:     fileType,
		SUID:         (mode & 0x0800) != 0,
		SGID:         (mode & 0x0400) != 0,
		Sticky:       (mode & 0x0200) != 0,
		AccessRights: mode & 0x01ff,
	}, nil
}

func (mode *Mode) Encode() uint16 {
	var suid, sgid, sticky uint16
	if mode.SUID {
		suid = 0x0800
	}
	if mode.SGID {
		sgid = 0x0400
	}
	if mode.Sticky {
		sticky = 0x0200
	}
	return mode.FileType.Encode() + suid + sgid + sticky + mode.AccessRights
}
