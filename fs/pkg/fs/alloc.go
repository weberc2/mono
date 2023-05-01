package fs

const (
	OutOfInodesErr constErr = "out of inodes"
	OutOfBlocksErr constErr = "out of blocks"
)

func AllocInode(fs *FileSystem) (Ino, error) {
	// TODO: mark the inode bitmap dirty
	ino, ok := fs.Descriptor.InodeBitmap.Alloc()
	if !ok {
		return InoOutOfInodes, OutOfInodesErr
	}
	fs.Superblock.FreeInodes--
	// TODO: mark the superblock dirty
	return Ino(ino) + InoFirst, nil
}

func AllocBlock(fs *FileSystem) (Block, error) {
	// TODO: mark the block bitmap dirty
	block, ok := fs.Descriptor.BlockBitmap.Alloc()
	if !ok {
		return 0, OutOfInodesErr
	}
	fs.Superblock.FreeBlocks--
	return fs.Superblock.FirstDataBlock() + Block(block), nil
}

func DeallocInode(fs *FileSystem, ino Ino) {
	fs.Superblock.FreeInodes++
	// TODO: mark superblock dirty
	fs.Descriptor.InodeBitmap.Free(uint64(ino))
	// TODO: mark inode bitmap dirty
}

func DeallocBlock(fs *FileSystem, block Block) {
	fs.Superblock.FreeBlocks++
	// TODO: mark superblock dirty
	fs.Descriptor.BlockBitmap.Free(uint64(block))
	// TODO: mark block bitmap dirty
}

// pub fn dealloc_block(fs: &mut Filesystem, block: u64) -> Result<()> {
// 	let (group_idx, local_idx) = get_block_group(fs, block);
// 	let group_id = group_idx as usize;
// 	let (local_byte, local_bit) = (local_idx / 8, local_idx % 8);
// 	fs.groups[group_id].desc.free_blocks_count += 1;
// 	fs.groups[group_id].block_bitmap[local_byte as usize] &= !(1 << local_bit);
// 	fs.groups[group_id].dirty = true;
// 	fs.superblock.free_blocks_count += 1;
// 	fs.superblock_dirty = true;
// 	Ok(())
//   }
