package fs

import (
	"fmt"
	"log"
)

const (
	BlockOutOfRange      Block = 0
	DirectBlocksPerInode Block = 12

	BlockOutOfRangeErr constErr = "block out of range"
)

func ReadInodeBlock(
	fs *FileSystem,
	inode *Inode,
	inodeBlock Block,
	offset Byte,
	p []byte,
) error {
	if offset+Byte(len(p)) > fs.Superblock.BlockSize {
		panic(fmt.Sprintf(
			"offset `%d` + len(p) `%d` must be less than or equal to the "+
				"block size `%d`",
			offset,
			len(p),
			fs.Superblock.BlockSize,
		))
	}

	realBlock, ok, err := BlockFromInodeBlock(fs, inode, inodeBlock)
	if err != nil {
		return fmt.Errorf(
			"reading inode `%d` block `%d` at offset `%d`: %w",
			inode.Ino,
			inodeBlock,
			offset,
			err,
		)
	}

	if !ok {
		return fmt.Errorf(
			"reading inode `%d` block `%d` at offset `%d`: %w",
			inode.Ino,
			inodeBlock,
			offset,
			BlockOutOfRangeErr,
		)
	}

	blockOffset := Byte(realBlock)*fs.Superblock.BlockSize + offset
	if err := ReadAt(fs.Volume, blockOffset, p); err != nil {
		return fmt.Errorf(
			"reading inode `%d` block `%d` at offset `%d`: %w",
			inode.Ino,
			inodeBlock,
			offset,
			err,
		)
	}
	return nil
}

func WriteInodeBlock(
	fs *FileSystem,
	inode *Inode,
	inodeBlock Block,
	offset Byte,
	buf []byte,
) error {
	if offset+Byte(len(buf)) > fs.Superblock.BlockSize {
		panic(fmt.Sprintf(
			"`offset (%d) + len(buf) (%d)` must be less than or equal to "+
				"the block size (%d)",
			offset,
			len(buf),
			fs.Superblock.BlockSize,
		))
	}

	block, ok, err := BlockFromInodeBlock(fs, inode, inodeBlock)
	if err != nil {
		return fmt.Errorf(
			"writing inode `%d` block `%d`: %w",
			inode.Ino,
			inodeBlock,
			err,
		)
	}
	if !ok {
		block, err = AllocInodeBlock(fs, inode)
		if err != nil {
			return fmt.Errorf(
				"writing inode `%d` block `%d`: %w",
				inode.Ino,
				inodeBlock,
				err,
			)
		}
		log.Printf("allocated block `%d`", block)
	}
	blockOffset := Byte(block)*fs.Superblock.BlockSize + offset
	if err := WriteAt(fs.Volume, blockOffset, buf); err != nil {
		return fmt.Errorf(
			"writing inode `%d` block `%d`: %w",
			inode.Ino,
			inodeBlock,
			err,
		)
	}

	return nil
}

func SetInodeBlockDirect(
	fs *FileSystem,
	inode *Inode,
	directIndex BlockListIndex,
	block Block,
) error {
	inode.DirectBlocks[directIndex] = block
	if err := UpdateInode(fs, inode); err != nil {
		return fmt.Errorf(
			"setting inode `%d` direct block `%d` to physical block `%d`: %w",
			inode.Ino,
			directIndex,
			err,
		)
	}
	return nil
}

func SetBlockSinglyIndirect(
	fs *FileSystem,
	singlyIndirectBlock func() (Block, error),
	singlyIndirectIndex BlockListIndex,
	block Block,
) error {
	indirectBlock, err := singlyIndirectBlock()
	if err != nil {
		return fmt.Errorf(
			"setting singly indirect block index `%d` to physical "+
				"block `%d`: %w",
			singlyIndirectIndex,
			block,
			fmt.Errorf("allocating singly indirect block: %w", err),
		)
	}
	if err := WriteIndirect(
		fs,
		indirectBlock,
		singlyIndirectIndex,
		block,
	); err != nil {
		return fmt.Errorf(
			"setting singly indirect block `%d` index `%d` to physical block "+
				"`%d`: %w",
			indirectBlock,
			singlyIndirectIndex,
			block,
			err,
		)
	}
	return nil
}

func SetInodeBlockDoublyIndirect(
	fs *FileSystem,
	inode *Inode,
	doublyIndirectBlock Block,
	doublyIndirectIndex BlockListIndex,
	singlyIndirectIndex BlockListIndex,
	block Block,
) (Block, error) {
	var singlyIndirectBlock Block
	var err error
	if doublyIndirectBlock == BlockOutOfRange {
		doublyIndirectBlock, err = AllocInodeBlock(fs, inode)
		if err != nil {
			return doublyIndirectBlock, fmt.Errorf(
				"setting doubly indirect block `%d` index `%d` to physical "+
					"block `%d`: %w",
				doublyIndirectBlock,
				doublyIndirectIndex,
				block,
				fmt.Errorf("allocating doubly indirect block: %w", err),
			)
		}
	} else {
		// if we didn't allocate a new block, then read the singly indirect
		// block. It may also be zero, but it will certainly be zero (or junk)
		// if it was freshly allocated--we'll handle the zero case outside of
		// this conditional.
		singlyIndirectBlock, err = ReadIndirect(
			fs,
			doublyIndirectBlock,
			doublyIndirectIndex,
		)
		if err != nil {
			return doublyIndirectBlock, fmt.Errorf(
				"setting doubly indirect block `%d` index `%d` to physical "+
					"block `%d`: discovering singly indirect block: %w",
				doublyIndirectBlock,
				doublyIndirectIndex,
				block,
				fmt.Errorf("allocating doubly indirect block: %w", err),
			)
		}
	}
	if err := SetBlockSinglyIndirect(
		fs,
		func() (Block, error) {
			if singlyIndirectBlock == BlockOutOfRange {
				singlyIndirectBlock, err := AllocBlock(fs)
				if err != nil {
					return BlockOutOfRange, fmt.Errorf(
						"allocating singly indirect block: %w",
						err,
					)
				}
				if err := WriteIndirect(
					fs,
					doublyIndirectBlock,
					doublyIndirectIndex,
					singlyIndirectBlock,
				); err != nil {
					DeallocBlock(fs, singlyIndirectBlock)
					return BlockOutOfRange, fmt.Errorf(
						"allocating singly indirect block: "+
							"updating doubly-indirect pointer table: %w",
						err,
					)
				}
				clone := *inode
				clone.Size += fs.Superblock.BlockSize
				if err := UpdateInode(fs, &clone); err != nil {
					DeallocBlock(fs, singlyIndirectBlock)
					return BlockOutOfRange, fmt.Errorf(
						"allocating singly indirect block: updating inode "+
							"size: %w",
						err,
					)
				}
				*inode = clone
			}
			return singlyIndirectBlock, nil
		},
		singlyIndirectIndex,
		block,
	); err != nil {
		return doublyIndirectBlock, fmt.Errorf(
			"setting doubly indirect block `%d` index `%d` to physical "+
				"block `%d`: %w",
			doublyIndirectBlock,
			doublyIndirectIndex,
			block,
			fmt.Errorf("allocating doubly indirect block: %w", err),
		)
	}
	return nil
}

func SetInodeBlockTriplyIndirect(
	fs *FileSystem,
	inode *Inode,
	triplyIndirectBlock Block,
	triplyIndirectIndex BlockListIndex,
	doublyIndirectIndex BlockListIndex,
	singlyIndirectIndex BlockListIndex,
	block Block,
) (Block, error) {
}

func SetInodeBlock(
	fs *FileSystem,
	inode *Inode,
	inodeBlock Block,
	block Block,
) error {
	// Expensive sanity check
	prevBlock, ok, err := BlockFromInodeBlock(fs, inode, inodeBlock)
	if err != nil {
		return fmt.Errorf(
			"setting inode `%d` block `%d` to physical block `%d`: "+
				"getting previous physical block for inode block: %w",
			inode.Ino,
			inodeBlock,
			block,
			err,
		)
	}
	if ok {
		panic(fmt.Sprintf(
			"setting inode `%d` block `%d` to physical block `%d`: "+
				"inode block is already mapped to physical block `%d`!",
			inode.Ino,
			inodeBlock,
			block,
			prevBlock,
		))
	}

	// direct: block pointer
	// singly indirect pointer: pointer to a list of block pointers
	// doubly indirect pointer: pointer to a list of singly indirect pointers
	// triply indirect pointer: pointer to a list of doubly indirect pointers
	pos := BlockPosFromInodeBlock(fs.Superblock.BlockSize, inodeBlock)
	if pos.Indirection == InodeBlockDirect {
		if err := SetInodeBlockDirect(
			fs,
			inode,
			pos.DirectIndex,
			block,
		); err != nil {
			return fmt.Errorf(
				"setting inode `%d` block `%d` to physical block `%d`: %w",
				inode.Ino,
				inodeBlock,
				block,
				err,
			)
		}
	} else if pos.Indirection == InodeBlockSinglyIndirect {
		err = SetBlockSinglyIndirect(
			fs,
			func() (Block, error) {
				if inode.SinglyIndirectBlock == BlockOutOfRange {
					singlyIndirectBlock, err := AllocBlock(fs)
					if err != nil {
						return BlockOutOfRange, fmt.Errorf(
							"allocating singly indirect block: %w",
							err,
						)
					}
					clone := *inode
					clone.SinglyIndirectBlock = singlyIndirectBlock
					clone.Size += fs.Superblock.BlockSize
					if err := UpdateInode(fs, &clone); err != nil {
						// if we can't update the inode, then free the
						// allocated block and
						DeallocBlock(fs, inode.SinglyIndirectBlock)
						return BlockOutOfRange, fmt.Errorf(
							"allocating singly indirect block: %w",
							err,
						)
					}
					// with the inode successfully committed to cache, update
					// the inode pointee
					*inode = clone
				}
				return inode.SinglyIndirectBlock, nil
			},
			pos.SinglyIndirectIndex,
			block,
		)
		if err != nil {
			return fmt.Errorf(
				"setting inode `%d` block `%d` to physical block `%d`: %w",
				inode.Ino,
				inodeBlock,
				block,
				err,
			)
		}
		return nil
	} else if pos.Indirection == InodeBlockDoublyIndirect {
		inode.DoublyIndirectBlock, err = SetInodeBlockDoublyIndirect(
			fs,
			inode,
			inode.DoublyIndirectBlock,
			pos.DoublyIndirectIndex,
			pos.SinglyIndirectIndex,
			block,
		)
		if err != nil {
			return fmt.Errorf(
				"setting inode `%d` block `%d` to physical block `%d`: %w",
				inode.Ino,
				inodeBlock,
				block,
				err,
			)
		}
		if err := UpdateInode(fs, inode); err != nil {
			return fmt.Errorf(
				"setting inode `%d` block `%d` to physical block `%d`: %w",
				inode.Ino,
				inodeBlock,
				block,
				err,
			)
		}
		return nil
	} else if pos.Indirection == InodeBlockTriplyIndirect {
		inode.TriplyIndirectBlock, err = SetInodeBlockTriplyIndirect(
			fs,
			inode,
			inode.TriplyIndirectBlock,
			pos.TriplyIndirectIndex,
			pos.DoublyIndirectIndex,
			pos.SinglyIndirectIndex,
			block,
		)
		if err != nil {
			return fmt.Errorf(
				"setting inode `%d` block `%d` to physical block `%d`: %w",
				inode.Ino,
				inodeBlock,
				block,
				err,
			)
		}
		if err := UpdateInode(fs, inode); err != nil {
			return fmt.Errorf(
				"setting inode `%d` block `%d` to physical block `%d`: %w",
				inode.Ino,
				inodeBlock,
				block,
				err,
			)
		}
		return nil
	}

	inodeIndirect := func(
		fs *FileSystem,
		inode *Inode,
		idx BlockListIndex,
	) (Block, error) {
		var block *Block
		if idx < BlockListIndex(DirectBlocksPerInode) {
			block = &inode.DirectBlocks[idx]
		} else if idx == BlockListIndex(DirectBlocksPerInode) {
			block = &inode.SinglyIndirectBlock
		} else if idx == BlockListIndex(DirectBlocksPerInode)+1 {
			block = &inode.DoublyIndirectBlock
		} else if idx == BlockListIndex(DirectBlocksPerInode)+2 {
			block = &inode.TriplyIndirectBlock
		}
		if *block == BlockOutOfRange {
			var err error
			*block, err = AllocIndirectBlock(fs, inode)
			if err != nil {
				return BlockOutOfRange, err
			}
			if err := UpdateInode(fs, inode); err != nil {
				return BlockOutOfRange, err
			}
		}
		return *block, nil
	}

	blockIndirect := func(
		fs *FileSystem,
		inode *Inode,
		indirect Block,
		entry BlockListIndex,
	) (Block, error) {
		oldBlock, err := ReadIndirect(fs, indirect, entry)
		if err != nil {
			return BlockOutOfRange, err
		}
		if oldBlock == 0 {
			newBlock, err := AllocIndirectBlock(fs, inode)
			if err != nil {
				return BlockOutOfRange, err
			}
			return newBlock, WriteIndirect(
				fs,
				indirect,
				entry,
				newBlock,
			)
		}
		return oldBlock, nil
	}

	pos := BlockPosFromInodeBlock(fs.Superblock.BlockSize, inodeBlock)
	switch pos.Indirection {
	case InodeBlockDirect:
		inode.DirectBlocks[pos.DirectIndex] = block
		if err := UpdateInode(fs, inode); err != nil {
			return fmt.Errorf(
				"setting inode `%d` block `%d` to physical block `%d`: %w",
				inode.Ino,
				inodeBlock,
				err,
			)
		}
	case InodeBlockSinglyIndirect:
		singlyIndirectBlock, err := inodeIndirect(
			fs,
			inode,
			BlockListIndex(DirectBlocksPerInode),
		)
		if err != nil {
			return fmt.Errorf(
				"setting inode `%d` block `%d` to physical block `%d`: %w",
				inode.Ino,
				inodeBlock,
				err,
			)
		}
		if err := WriteIndirect(
			fs,
			singlyIndirectBlock,
			InodeBlockSinglyIndirect,
			block,
		); err != nil {
			return fmt.Errorf(
				"setting inode `%d` block `%d` to physical block `%d`: %w",
				inode.Ino,
				inodeBlock,
				err,
			)
		}
	case InodeBlockDoublyIndirect:
		doublyIndirectBlock, err := inodeIndirect(
			fs,
			inode,
			BlockListIndex(DirectBlocksPerInode+1),
		)
		if err != nil {
			return fmt.Errorf(
				"setting inode `%d` block `%d` to physical block `%d`: %w",
				inode.Ino,
				inodeBlock,
				err,
			)
		}
		singlyIndirectBlock, err := blockIndirect(fs, inode, doublyIndirectBlock)
	}
}

func AllocInodeBlock(fs *FileSystem, inode *Inode) (Block, error) {
	block, err := AllocBlock(fs)
	if err != nil {
		return 0, fmt.Errorf(
			"allocating block for inode `%d`: %w",
			inode.Ino,
			err,
		)
	}

	inode.Size += fs.Superblock.BlockSize

	if err := UpdateInode(fs, inode); err != nil {
		return 0, fmt.Errorf(
			"allocating block for inode `%d`: %w",
			inode.Ino,
			err,
		)
	}

	return block, nil
}

func BlockFromInodeBlock(
	fs *FileSystem,
	inode *Inode,
	inodeBlock Block,
) (Block, bool, error) {
	block, err := BlockFromBlockPos(
		fs,
		inode,
		BlockPosFromInodeBlock(fs.Superblock.BlockSize, inodeBlock),
	)
	if err != nil {
		return 0, false, err
	}
	return block, block != 0, err
}

func BlockFromBlockPos(fs *FileSystem, inode *Inode, pos BlockPos) (Block, error) {
	switch pos.Indirection {
	case InodeBlockDirect:
		return DirectBlock(inode, pos.DirectIndex), nil
	case InodeBlockSinglyIndirect:
		return SinglyIndirectBlock(
			fs,
			inode.SinglyIndirectBlock,
			pos.SinglyIndirectIndex,
		)
	case InodeBlockDoublyIndirect:
		return DoublyIndirectBlock(
			fs,
			inode.DoublyIndirectBlock,
			pos.SinglyIndirectIndex,
			pos.DoublyIndirectIndex,
		)
	case InodeBlockTriplyIndirect:
		return TriplyIndirectBlock(
			fs,
			inode.TriplyIndirectBlock,
			pos.SinglyIndirectIndex,
			pos.DoublyIndirectIndex,
			pos.TriplyIndirectIndex,
		)
	case InodeBlockOutOfRange:
		return 0, fmt.Errorf("exceeded max number of blocks in an inode")
	default:
		panic(fmt.Sprintf("invalid BlockPosLevel: %d", pos.Indirection))
	}
}

func DirectBlock(inode *Inode, directIndex BlockListIndex) Block {
	return inode.DirectBlocks[directIndex]
}

func SinglyIndirectBlock(
	fs *FileSystem,
	singlyIndirectBlock Block,
	singlyIndirectIndex BlockListIndex,
) (Block, error) {
	return ReadIndirect(fs, singlyIndirectBlock, singlyIndirectIndex)
}

func DoublyIndirectBlock(
	fs *FileSystem,
	doublyIndirectBlock Block,
	singlyIndirectIndex BlockListIndex,
	doublyIndirectIndex BlockListIndex,
) (Block, error) {
	singlyIndirectBlock, err := ReadIndirect(
		fs,
		doublyIndirectBlock,
		doublyIndirectIndex,
	)
	if err != nil {
		return 0, err
	}
	if singlyIndirectBlock == 0 {
		return 0, err
	}

	return SinglyIndirectBlock(fs, singlyIndirectBlock, singlyIndirectIndex)
}

func TriplyIndirectBlock(
	fs *FileSystem,
	triplyIndirectBlock Block,
	singlyIndirectIndex BlockListIndex,
	doublyIndirectIndex BlockListIndex,
	triplyIndirectIndex BlockListIndex,
) (Block, error) {
	doublyIndirectBlock, err := ReadIndirect(
		fs,
		triplyIndirectBlock,
		triplyIndirectIndex,
	)
	if err != nil {
		return 0, err
	}
	if doublyIndirectBlock == 0 {
		return 0, nil
	}
	return DoublyIndirectBlock(
		fs,
		doublyIndirectBlock,
		singlyIndirectIndex,
		doublyIndirectIndex,
	)
}

func ReadIndirect(fs *FileSystem, indirectBlock Block, index BlockListIndex) (Block, error) {
	pointersPerBlock := fs.Superblock.BlockPointersPerBlock()
	if debug && index >= BlockListIndex(pointersPerBlock) {
		panic(fmt.Sprintf(
			"entry `%d` exceeds the number of block pointers that can fit in "+
				"a block (`%d`)",
			index,
			pointersPerBlock,
		))
	}
	var buf [4]byte
	entryOffset := Byte(indirectBlock)*fs.Superblock.BlockSize + Byte(index*4)
	if err := ReadAt(fs.Volume, entryOffset, buf[:]); err != nil {
		return 0, fmt.Errorf(
			"reading indirect block `%d` at index `%d`",
			indirectBlock,
			index,
		)
	}
	return getBlock(buf[:]), nil
}
