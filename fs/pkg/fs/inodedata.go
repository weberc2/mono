package fs

import (
	"fmt"
	"log"
)

func ReadInodeData(
	fs *FileSystem,
	inode *Inode,
	offset Byte,
	p []byte,
) (Byte, error) {
	blockSize := fs.Superblock.BlockSize
	maxLength := Min(Byte(len(p)), inode.Size-offset)
	chunkBegin := Byte(0)

	for chunkBegin < maxLength {
		chunkBlock := Block((offset + chunkBegin) / blockSize)
		chunkOffset := (offset + chunkBegin) % blockSize
		chunkLength := Min(maxLength-chunkBegin, blockSize-chunkOffset)
		if err := ReadInodeBlock(
			fs,
			inode,
			chunkBlock,
			chunkOffset,
			p[chunkBegin:chunkLength],
		); err != nil {
			return 0, fmt.Errorf(
				"reading data for inode `%d` at offset `%d`: %w",
				inode.Ino,
				offset,
				err,
			)
		}
		chunkBegin += chunkLength
	}
	return chunkBegin, nil
}

func WriteInodeData(
	fs *FileSystem,
	inode *Inode,
	offset Byte,
	buf []byte,
) (Byte, error) {
	var blockSize Byte = fs.Superblock.BlockSize
	var chunkBegin Byte = 0

	for chunkBegin < Byte(len(buf)) {
		chunkBlock := Block((offset + chunkBegin) / blockSize)
		chunkOffset := (offset + chunkBegin) % blockSize
		chunkLength := Min(Byte(len(buf))-chunkBegin, blockSize-chunkOffset)
		log.Printf("writing inode block")
		if err := WriteInodeBlock(
			fs,
			inode,
			chunkBlock,
			chunkOffset,
			buf[chunkBegin:chunkLength],
		); err != nil {
			return 0, fmt.Errorf(
				"writing data for inode `%d`: %w",
				inode.Ino,
				err,
			)
		}
		chunkBegin += chunkLength
	}

	if inode.Size < offset+chunkBegin {
		inode.Size = offset + chunkBegin
		if err := UpdateInode(fs, inode); err != nil {
			return 0, fmt.Errorf(
				"writing data for inode `%d`: %w",
				inode.Ino,
				err,
			)
		}
	}

	return chunkBegin, nil
}
