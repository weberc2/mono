package fs

type BlockListIndex int

type InodeBlockDirection int

const (
	InodeBlockDirect InodeBlockDirection = iota
	InodeBlockSinglyIndirect
	InodeBlockDoublyIndirect
	InodeBlockTriplyIndirect
	InodeBlockOutOfRange
)

type BlockPos struct {
	Indirection         InodeBlockDirection
	DirectIndex         BlockListIndex
	SinglyIndirectIndex BlockListIndex
	DoublyIndirectIndex BlockListIndex
	TriplyIndirectIndex BlockListIndex
}

func NewDirectBlockPos(directIndex BlockListIndex) BlockPos {
	return BlockPos{Indirection: InodeBlockDirect, DirectIndex: directIndex}
}

func NewSinglyIndirectBlockPos(singlyIndirectIndex BlockListIndex) BlockPos {
	return BlockPos{
		Indirection:         InodeBlockSinglyIndirect,
		SinglyIndirectIndex: singlyIndirectIndex,
	}
}

func NewDoublyIndirectBlockPos(
	singlyIndirectIndex BlockListIndex,
	doublyIndirectIndex BlockListIndex,
) BlockPos {
	return BlockPos{
		Indirection:         InodeBlockDoublyIndirect,
		SinglyIndirectIndex: singlyIndirectIndex,
		DoublyIndirectIndex: doublyIndirectIndex,
	}
}

func NewTriplyIndirectBlockPos(
	singlyIndirectIndex BlockListIndex,
	doublyIndirectIndex BlockListIndex,
	triplyIndirectIndex BlockListIndex,
) BlockPos {
	return BlockPos{
		Indirection:         InodeBlockTriplyIndirect,
		SinglyIndirectIndex: singlyIndirectIndex,
		DoublyIndirectIndex: doublyIndirectIndex,
		TriplyIndirectIndex: triplyIndirectIndex,
	}
}

func NewOutOfRangeBlockPos() BlockPos {
	return BlockPos{Indirection: InodeBlockOutOfRange}
}

func BlockPosFromInodeBlock(blockSize Byte, inodeBlock Block) BlockPos {
	indirect1Size := Block(blockSize / BlockPointerSize)
	indirect2Size := Block(indirect1Size * indirect1Size)
	indirect3Size := Block(indirect1Size * indirect2Size)

	if inodeBlock < DirectBlocksPerInode {
		return NewDirectBlockPos(BlockListIndex(inodeBlock))
	} else if inodeBlock < DirectBlocksPerInode+indirect1Size {
		return NewSinglyIndirectBlockPos(
			BlockListIndex(inodeBlock - DirectBlocksPerInode),
		)
	} else if inodeBlock < DirectBlocksPerInode+indirect1Size+indirect2Size {
		base := inodeBlock - DirectBlocksPerInode - indirect1Size
		return NewDoublyIndirectBlockPos(
			BlockListIndex(base/indirect1Size),
			BlockListIndex(base%indirect1Size),
		)
	} else if inodeBlock < DirectBlocksPerInode+indirect1Size+indirect2Size+indirect3Size {
		base := inodeBlock - DirectBlocksPerInode - indirect1Size - indirect2Size
		return NewTriplyIndirectBlockPos(
			BlockListIndex(base/indirect2Size),
			BlockListIndex((base%indirect2Size)/indirect1Size),
			BlockListIndex((base%indirect2Size)%indirect1Size),
		)
	} else {
		return NewOutOfRangeBlockPos()
	}
}

// fn inode_block_to_pos(fs: &Filesystem, inode_block: u64) -> BlockPos {
// 	let indirect_1_size: u64 = fs.block_size() / 4;
// 	let indirect_2_size = indirect_1_size * indirect_1_size;
// 	let indirect_3_size = indirect_1_size * indirect_2_size;
// 	if inode_block < 12 {
// 	  BlockPos::Level0(inode_block)
// 	} else if inode_block < 12 + indirect_1_size {
// 	  BlockPos::Level1(inode_block - 12)
// 	} else if inode_block < 12 + indirect_1_size + indirect_2_size {
// 	  let base = inode_block - 12 - indirect_1_size;
// 	  BlockPos::Level2(base / indirect_1_size, base % indirect_1_size)
// 	} else if inode_block < 12 + indirect_1_size + indirect_2_size + indirect_3_size {
// 	  let base = inode_block - 12 - indirect_1_size - indirect_2_size;
// 	  BlockPos::Level3(base / indirect_2_size,
// 		(base % indirect_2_size) / indirect_1_size,
// 		(base % indirect_2_size) % indirect_1_size)
// 	} else {
// 	  BlockPos::OutOfRange
// 	}
//   }
