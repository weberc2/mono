package fs

import "encoding/binary"

const (
	SizeIno            Byte = Size64
	SizeByte           Byte = Size64
	SizeDirEntryHeader Byte = dirEntryFieldInoSize +
		dirEntryFieldRecLenSize +
		dirEntryFieldNameLenSize +
		dirEntryFieldFileTypeSize

	superblockFieldMagic uint16 = iota
	superblockFieldBlockSize
	superblockFieldBlockCount
	superblockFieldInodeCount
	superblockFieldFreeBlocks
	superblockFieldFreeInodes

	dirEntryFieldInoOffset      Byte = 0
	dirEntryFieldInoSize        Byte = SizeIno
	dirEntryFieldRecLenOffset   Byte = dirEntryFieldInoOffset + dirEntryFieldInoSize
	dirEntryFieldRecLenSize     Byte = SizeByte
	dirEntryFieldNameLenOffset  Byte = dirEntryFieldRecLenOffset + dirEntryFieldRecLenSize
	dirEntryFieldNameLenSize    Byte = 1
	dirEntryFieldFileTypeOffset Byte = dirEntryFieldNameLenOffset + dirEntryFieldNameLenSize
	dirEntryFieldFileTypeSize   Byte = 1
)

func EncodeSuperblock(superblock *Superblock, p *[SuperblockSize]byte) {
	putU64(p[superblockFieldMagic*size64:], SuperblockMagic)
	putByte(p[superblockFieldBlockSize*size64:], superblock.BlockSize)
	putBlock(p[superblockFieldBlockCount*size64:], superblock.BlockCount)
	putIno(p[superblockFieldInodeCount*size64:], superblock.InodeCount)
	putBlock(p[superblockFieldFreeBlocks*size64:], superblock.FreeBlocks)
	putIno(p[superblockFieldFreeInodes*size64:], superblock.FreeInodes)
}

func EncodeInode(inode *Inode, buf *[InodeSize]byte) {
	p := buf[:]
	encodeU32(p, inodeFieldAttrUID, inode.Attr.UID)
	encodeU32(p, inodeFieldAttrGID, inode.Attr.GID)
	encodeU32(p, inodeFieldAttrATime, inode.Attr.ATime)
	encodeU32(p, inodeFieldAttrCTime, inode.Attr.CTime)
	encodeU32(p, inodeFieldAttrMTime, inode.Attr.MTime)
	encodeU32(p, inodeFieldAttrDTime, inode.Attr.DTime)
	encodeU32(p, inodeFieldFlags, inode.Flags)
	encodeU32(p, inodeFieldACL, inode.ACL)
	encodeU16(p, inodeFieldMode, EncodeMode(&inode.Mode))
	encodeU16(p, inodeFieldLinksCount, inode.LinksCount)
	encodeU64(p, inodeFieldSize, uint64(inode.Size))

	directBlockOffset := inodeFieldOffsets[inodeFieldDirectBlocks]
	for i, block := range inode.DirectBlocks {
		putU64(p[directBlockOffset+uint16(i)*size64:], uint64(block))
	}

	encodeBlock(p, inodeFieldSinglyIndirectBlock, inode.SinglyIndirectBlock)
	encodeBlock(p, inodeFieldDoublyIndirectBlock, inode.DoublyIndirectBlock)
	encodeBlock(p, inodeFieldTriplyIndirectBlock, inode.TriplyIndirectBlock)
}

func EncodeMode(mode *Mode) uint16 {
	out := EncodeFileType(mode.Type)
	if mode.SUID {
		out += 0x0800
	}
	if mode.SGID {
		out += 0x0400
	}
	if mode.Sticky {
		out += 0x0200
	}
	return out + mode.AccessRights
}

func EncodeFileType(fileType FileType) uint16 {
	var out uint16
	switch fileType {
	case FileTypeFifo:
		out = 1
	case FileTypeCharDev:
		out = 2
	case FileTypeDir:
		out = 4
	case FileTypeBlockDev:
		out = 6
	case FileTypeRegular:
		out = 8
	case FileTypeSymlink:
		out = 10
	case FileTypeSocket:
		out = 12
	}
	return out << 12
}

func EncodeDirEntryHeader(header *DirEntryHeader, p *[SizeDirEntryHeader]byte) {
	putIno(p[dirEntryFieldInoOffset:], header.Ino)
	putByte(p[dirEntryFieldRecLenOffset:], header.RecLen)
	p[dirEntryFieldNameLenOffset] = header.NameLen
	p[dirEntryFieldFileTypeOffset] = byte(header.FileType)
}

func encodeU16(p []byte, field inodeField, u uint16) {
	putU16(p[inodeFieldOffsets[field]:], u)
}

func encodeU32(p []byte, field inodeField, u uint32) {
	putU32(p[inodeFieldOffsets[field]:], u)
}

func encodeBlock(p []byte, field inodeField, u Block) {
	encodeU64(p, field, uint64(u))
}

func encodeU64(p []byte, field inodeField, u uint64) {
	putU64(p[inodeFieldOffsets[field]:], u)
}

func putU16(p []byte, u uint16) {
	binary.BigEndian.PutUint16(p, u)
}

func putU32(p []byte, u uint32) {
	binary.BigEndian.PutUint32(p, u)
}

func putIno(p []byte, u Ino) {
	putU64(p, uint64(u))
}

func putBlock(p []byte, u Block) {
	putU64(p, uint64(u))
}

func putByte(p []byte, u Byte) {
	putU64(p, uint64(u))
}

func putU64(p []byte, u uint64) {
	binary.BigEndian.PutUint64(p, u)
}
