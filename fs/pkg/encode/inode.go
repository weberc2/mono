package encode

import (
	"fmt"

	. "github.com/weberc2/mono/fs/pkg/types"
)

func EncodeInode(inode *Inode, b *[InodeSize]byte) {
	p := b[:]

	putU8(p, inodeFileTypeStart, uint8(inode.FileType))
	putBytePointer(p, inodeSizeStart, inode.Size)
	putU16(p, inodeLinksCountStart, inode.LinksCount)

	for i := Byte(0); i < Byte(DirectBlocksCount); i++ {
		blockPointerStart := inodeDirectBlocksStart + i*BlockPointerSize
		EncodeBlock(
			inode.DirectBlocks[i],
			(*[BlockPointerSize]byte)(p[blockPointerStart:]),
		)
	}

	EncodeBlock(
		inode.SinglyIndirectBlock,
		(*[BlockPointerSize]byte)(p[inodeSinglyIndStart:inodeSinglyIndEnd]),
	)

	EncodeBlock(
		inode.DoublyIndirectBlock,
		(*[BlockPointerSize]byte)(p[inodeDoublyIndStart:inodeDoublyIndEnd]),
	)

	EncodeBlock(
		inode.TriplyIndirectBlock,
		(*[BlockPointerSize]byte)(p[inodeTriplyIndStart:inodeTriplyIndEnd]),
	)
}

func DecodeInode(inode *Inode, b *[InodeSize]byte) error {
	p := b[:]

	// store this in a temporary until we've validated it; we strongly prefer
	// to avoid mutating the `inode` pointee until we're sure that no errors
	// will be returned.
	ft := FileType(getU8(p, inodeFileTypeStart))
	if err := ft.Validate(); err != nil {
		return fmt.Errorf("decoding inode: %w", err)
	}

	inode.FileType = ft
	inode.Size = getBytePointer(p, inodeSizeStart)
	inode.LinksCount = getU16(p, inodeLinksCountStart)

	for i := Byte(0); i < Byte(DirectBlocksCount); i++ {
		blockPointerStart := inodeDirectBlocksStart + i*BlockPointerSize
		inode.DirectBlocks[i] = DecodeBlock(
			(*[BlockPointerSize]byte)(p[blockPointerStart:]),
		)
	}

	inode.SinglyIndirectBlock = DecodeBlock(
		(*[BlockPointerSize]byte)(p[inodeSinglyIndStart:inodeSinglyIndEnd]),
	)

	inode.DoublyIndirectBlock = DecodeBlock(
		(*[BlockPointerSize]byte)(p[inodeDoublyIndStart:inodeDoublyIndEnd]),
	)

	inode.TriplyIndirectBlock = DecodeBlock(
		(*[BlockPointerSize]byte)(p[inodeTriplyIndStart:inodeTriplyIndEnd]),
	)

	return nil
}

const (
	inodeFileTypeStart = 0
	inodeFileTypeSize  = 1
	inodeFileTypeEnd   = inodeFileTypeStart + inodeFileTypeSize

	inodeSizeStart = inodeFileTypeEnd
	inodeSizeSize  = BytePointerSize
	inodeSizeEnd   = inodeSizeStart + inodeSizeSize

	inodeLinksCountStart = inodeSizeEnd
	inodeLinksCountSize  = 2
	inodeLinksCountEnd   = inodeLinksCountStart + inodeLinksCountSize

	inodeDirectBlocksStart = inodeLinksCountEnd
	inodeDirectBlocksSize  = Byte(DirectBlocksCount) * BlockPointerSize
	inodeDirectBlocksEnd   = inodeDirectBlocksStart + inodeDirectBlocksSize

	inodeSinglyIndStart = inodeDirectBlocksEnd
	inodeSinglyIndSize  = BlockPointerSize
	inodeSinglyIndEnd   = inodeSinglyIndStart + inodeSinglyIndSize

	inodeDoublyIndStart = inodeSinglyIndEnd
	inodeDoublyIndSize  = BlockPointerSize
	inodeDoublyIndEnd   = inodeDoublyIndStart + inodeDoublyIndSize

	inodeTriplyIndStart = inodeDoublyIndEnd
	inodeTriplyIndSize  = BlockPointerSize
	inodeTriplyIndEnd   = inodeTriplyIndStart + inodeTriplyIndSize
)
