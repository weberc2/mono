package fs

type Block uint64

type Byte int64

const (
	Size64                      Byte   = 8
	DefaultBlockSize            Byte   = 1024
	BlockPointerSize            Byte   = Size64 // 64-bit block pointers
	SuperblockOffset            Byte   = 1024
	SuperblockSize              Byte   = DefaultBlockSize
	SuperblockMagic             uint64 = 7234295460216005990 // ascii "deadbeef"
	DescriptorUsedDirsCountSize Byte   = Size64
	InodeSize                   Byte   = 1024
)

type Superblock struct {
	BlockSize  Byte
	BlockCount Block
	InodeCount Ino
	FreeBlocks Block
	FreeInodes Ino
	// TODO: either move the `Descriptor.UsedDirsCount` counter here with the
	// other counters or move the other counters into the `Descriptor` struct.
}

func NewSuperblock(blockSize Byte, blocks Block, inodes Ino) Superblock {
	return Superblock{
		BlockSize:  blockSize,
		BlockCount: blocks,
		InodeCount: inodes,
		FreeBlocks: blocks - FirstDataBlock(blockSize, blocks, inodes),
		FreeInodes: inodes - InoFirst + 1,
	}
}

func (superblock *Superblock) BlockPointersPerBlock() Block {
	return Block(superblock.BlockSize / BlockPointerSize)
}

func (superblock *Superblock) DescriptorOffset() Byte {
	return DescriptorOffset(superblock.BlockSize)
}

func (superblock *Superblock) DescriptorSize() Byte {
	return DescriptorSize(superblock.BlockCount, superblock.InodeCount)
}

func DescriptorSize(blockCount Block, inodeCount Ino) Byte {
	return Size64 + BlockBitmapSize(blockCount) + InodeBitmapSize(inodeCount)
}

func (superblock *Superblock) BlockBitmapOffset() Byte {
	return BlockBitmapOffset(superblock.BlockSize)
}

func BlockBitmapOffset(blockSize Byte) Byte {
	return DescriptorUsedDirsCountSize + DescriptorOffset(blockSize)
}

func DescriptorOffset(blockSize Byte) Byte {
	superblockStartBlock := DivCiel(SuperblockOffset, blockSize)
	superblockBlocks := DivCiel(SuperblockSize, blockSize)
	return (superblockStartBlock + superblockBlocks) * blockSize
}

func (superblock *Superblock) BlockBitmapSize() Byte {
	return BlockBitmapSize(superblock.BlockCount)
}

func BlockBitmapSize(blockCount Block) Byte {
	return Byte(DivCiel(blockCount, 8))
}

func (superblock *Superblock) InodeBitmapOffset() Byte {
	return InodeBitmapOffset(superblock.BlockSize, superblock.BlockCount)
}

func InodeBitmapOffset(blockSize Byte, blockCount Block) Byte {
	return BlockBitmapOffset(blockSize) + BlockBitmapSize(blockCount)
}

func (superblock *Superblock) InodeBitmapSize() Byte {
	return InodeBitmapSize(superblock.InodeCount)
}

func InodeBitmapSize(inodeCount Ino) Byte {
	return Byte(DivCiel(inodeCount, 8))
}

func (superblock *Superblock) InodeTableOffset() Byte {
	return InodeTableOffset(
		superblock.BlockSize,
		superblock.BlockCount,
		superblock.InodeCount,
	)
}

func InodeTableOffset(blockSize Byte, blockCount Block, inodeCount Ino) Byte {
	return DescriptorOffset(blockSize) + DescriptorSize(blockCount, inodeCount)
}

func (superblock *Superblock) InodeOffset(ino Ino) Byte {
	return superblock.InodeTableOffset() + ino.TableOffset()
}

func (superblock *Superblock) InodeTableSize() Byte {
	return InodeTableSize(superblock.InodeCount)
}

func InodeTableSize(inodeCount Ino) Byte {
	return Byte(inodeCount) * InodeSize
}

func (superblock *Superblock) FirstDataBlock() Block {
	return FirstDataBlock(
		superblock.BlockSize,
		superblock.BlockCount,
		superblock.InodeCount,
	)
}

func FirstDataBlock(blockSize Byte, blockCount Block, inodeCount Ino) Block {
	return Block(DivCiel(
		InodeTableOffset(blockSize, blockCount, inodeCount)+
			InodeTableSize(inodeCount),
		blockSize,
	))
}
