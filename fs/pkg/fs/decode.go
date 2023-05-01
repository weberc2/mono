package fs

import (
	"encoding/binary"
	"fmt"
)

const (
	BadMagicErr constErr = "bad magic"
)

func DecodeSuperblock(superblock *Superblock, p *[SuperblockSize]byte) error {
	if magic := getU64(
		p[superblockFieldMagic*size64:],
	); magic != SuperblockMagic {
		return fmt.Errorf(
			"decoding superblock: decoded magic `%#x`: %w",
			magic,
			BadMagicErr,
		)
	}
	*superblock = Superblock{
		BlockSize:  getByte(p[superblockFieldBlockSize*size64:]),
		BlockCount: getBlock(p[superblockFieldBlockCount*size64:]),
		InodeCount: getIno(p[superblockFieldInodeCount*size64:]),
		FreeBlocks: getBlock(p[superblockFieldFreeBlocks*size64:]),
		FreeInodes: getIno(p[superblockFieldFreeInodes*size64:]),
	}
	return nil
}

// DecodeInode populates `inode` with data from `buf`. Note that `inode.Ino` is
// not populated because the ino isn't discernible from an encoded inode.
func DecodeInode(inode *Inode, buf *[InodeSize]byte) error {
	p := buf[:]
	*inode = Inode{
		Attr: FileAttr{
			UID:   decodeU32(p, inodeFieldAttrUID),
			GID:   decodeU32(p, inodeFieldAttrGID),
			ATime: decodeU32(p, inodeFieldAttrATime),
			CTime: decodeU32(p, inodeFieldAttrCTime),
			MTime: decodeU32(p, inodeFieldAttrMTime),
			DTime: decodeU32(p, inodeFieldAttrDTime),
		},
		Flags:               decodeU32(p, inodeFieldFlags),
		ACL:                 decodeU32(p, inodeFieldACL),
		LinksCount:          decodeU16(p, inodeFieldLinksCount),
		Size:                decodeByte(p, inodeFieldSize),
		SinglyIndirectBlock: decodeBlock(p, inodeFieldSinglyIndirectBlock),
		DoublyIndirectBlock: decodeBlock(p, inodeFieldDoublyIndirectBlock),
		TriplyIndirectBlock: decodeBlock(p, inodeFieldTriplyIndirectBlock),
	}
	directBlocksOffset := inodeFieldOffsets[inodeFieldDirectBlocks]
	for i := range inode.DirectBlocks {
		inode.DirectBlocks[i] = getBlock(
			p[directBlocksOffset+uint16(i)*size64:],
		)
	}
	return DecodeMode(&inode.Mode, decodeU16(p, inodeFieldMode))
}

func DecodeMode(m *Mode, mode uint16) error {
	typeNibble := mode & 0xf000 >> 12
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
		fmt.Errorf(
			"decoding mode `%#x`: unknown file type `%d`",
			mode,
			typeNibble,
		)
	}

	*m = Mode{
		Type:         fileType,
		SUID:         mode&0x0800 != 0,
		SGID:         mode&0x0400 != 0,
		Sticky:       mode&0x0200 != 0,
		AccessRights: mode & 0x01ff,
	}
	return nil
}

func DecodeDirEntryHeader(header *DirEntryHeader, p *[SizeDirEntryHeader]byte) {
	header.Ino = getIno(p[dirEntryFieldInoOffset:])
	header.RecLen = getByte(p[dirEntryFieldRecLenOffset:])
	header.NameLen = p[dirEntryFieldNameLenOffset]
	header.FileType = FileType(p[dirEntryFieldFileTypeOffset])
}

func decodeU16(p []byte, field inodeField) uint16 {
	return getU16(p[inodeFieldOffsets[field]:])
}

func decodeU32(p []byte, field inodeField) uint32 {
	return getU32(p[inodeFieldOffsets[field]:])
}

func decodeByte(p []byte, field inodeField) Byte {
	return Byte(decodeU64(p, field))
}

func decodeBlock(p []byte, field inodeField) Block {
	return Block(decodeU64(p, field))
}

func decodeU64(p []byte, field inodeField) uint64 {
	return getU64(p[inodeFieldOffsets[field]:])
}

func getU16(p []byte) uint16 {
	return binary.BigEndian.Uint16(p)
}

func getU32(p []byte) uint32 {
	return binary.BigEndian.Uint32(p)
}

func getIno(p []byte) Ino {
	return Ino(getU64(p))
}

func getBlock(p []byte) Block {
	return Block(getU64(p))
}

func getByte(p []byte) Byte {
	return Byte(getU64(p))
}

func getU64(p []byte) uint64 {
	return binary.BigEndian.Uint64(p)
}
