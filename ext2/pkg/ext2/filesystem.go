package ext2

import (
	"errors"
	"fmt"
)

const RootIno Ino = 2

type FileSystem struct {
	Volume          Volume
	Superblock      Superblock
	SuperblockBytes *[SuperblockSize]byte
	SuperblockDirty bool
	Groups          []Group
	InodeCache      map[Ino]Inode
	DirtyInos       map[Ino]struct{}
	ReusedInos      map[Ino]struct{}
	CacheQueue      Ring
}

func (fs *FileSystem) BlockSize() uint64 {
	return 1024 << fs.Superblock.LogBlockSize
}

func (fs *FileSystem) GroupCount() GroupID {
	a := GroupID(fs.Superblock.BlocksCount)
	b := GroupID(fs.Superblock.BlocksPerGroup)
	return (a + b - 1) / b
}

func (fs *FileSystem) Mount(volume Volume) error {
	var superblockBytes [1024]byte
	if err := volume.Read(1024, superblockBytes[:]); err != nil {
		return fmt.Errorf("mounting filesystem: %w", err)
	}

	sb, err := DecodeSuperblock(&superblockBytes, false)
	if err != nil {
		return fmt.Errorf("mounting filesystem: %w", err)
	}

	tmp := FileSystem{
		Volume:          volume,
		Superblock:      sb,
		SuperblockBytes: &superblockBytes,
		SuperblockDirty: false,
		Groups:          nil,
		InodeCache:      map[Ino]Inode{},
		DirtyInos:       map[Ino]struct{}{},
		ReusedInos:      map[Ino]struct{}{},
		CacheQueue:      NewRing(), // empty ring
	}

	tmp.Groups = make([]Group, tmp.GroupCount())
	for i := GroupID(0); i < GroupID(tmp.GroupCount()); i++ {
		group, err := fs.ReadGroup(i)
		if err != nil {
			return fmt.Errorf("mounting filesystem: %w", err)
		}
		tmp.Groups[i] = group
	}

	if err := fs.FlushSuperblock(false); err != nil {
		return fmt.Errorf("mounting filesystem: %w", err)
	}

	return nil
}

func (fs *FileSystem) ReadGroup(groupID GroupID) (Group, error) {
	tableBlock := uint64(fs.Superblock.FirstDataBlock) + 1
	desc, err := fs.ReadGroupDesc(tableBlock, groupID)
	if err != nil {
		return Group{}, fmt.Errorf("reading group `%#x`: %w", groupID, err)
	}

	blockBitmapOffset := uint64(desc.BlockBitmap) * fs.BlockSize()
	blockBitmap := make([]byte, uint64(fs.Superblock.BlocksPerGroup)/8)
	if err := fs.Volume.Read(blockBitmapOffset, blockBitmap); err != nil {
		return Group{}, fmt.Errorf(
			"reading group `%#x`: reading block bitmap: %w",
			groupID,
			err,
		)
	}

	inodeBitmapOffset := uint64(desc.InodeBitmap) * fs.BlockSize()
	inodeBitmap := make([]byte, uint64(fs.Superblock.InodesPerGroup)/8)
	if err := fs.Volume.Read(inodeBitmapOffset, inodeBitmap); err != nil {
		return Group{}, fmt.Errorf(
			"reading group `%#x`: reading inode bitmap: %w",
			groupID,
			err,
		)
	}

	return Group{
		Idx:         groupID,
		Desc:        desc,
		BlockBitmap: blockBitmap,
		InodeBitmap: inodeBitmap,
		Dirty:       false,
	}, nil
}

func (fs *FileSystem) ReadGroupDesc(
	tableBlock uint64,
	groupID GroupID,
) (GroupDesc, error) {
	offset := tableBlock + fs.BlockSize() + uint64(groupID)*32
	var descBuf [32]byte
	if err := fs.Volume.Read(offset, descBuf[:]); err != nil {
		return GroupDesc{}, fmt.Errorf(
			"reading descriptor for group `%#x` in table block `%#x`: %w",
			groupID,
			tableBlock,
			err,
		)
	}
	return DecodeGroupDesc(&descBuf), nil
}

func (fs *FileSystem) FlushSuperblock(clean bool) error {
	state := StateClean
	if !clean {
		state = StateDirty
	}
	fs.SuperblockDirty = fs.SuperblockDirty || fs.Superblock.State != state
	fs.Superblock.State = state

	if fs.SuperblockDirty {
		fs.Superblock.Encode(fs.SuperblockBytes)

		if err := fs.Volume.Write(1024, fs.SuperblockBytes[:]); err != nil {
			return fmt.Errorf("flushing superblock: %w", err)
		}

		fs.SuperblockDirty = false
	}

	return nil
}

func (fs *FileSystem) GetInode(ino Ino) (Inode, error) {
	inode, found := fs.InodeCache[ino]
	if found {
		fs.ReusedInos[ino] = struct{}{}
		return inode, nil
	}

	inode, err := fs.ReadInode(ino)
	if err != nil {
		return Inode{}, fmt.Errorf("fetching inode `%#x`: %w", ino, err)
	}

	fs.InodeCache[ino] = inode
	fs.CacheQueue.PushBack(ino)
	if err := fs.RefitInodeCache(); err != nil {
		return Inode{}, fmt.Errorf("fetching inode `%#x`: %w", ino, err)
	}
	return inode, nil
}

func (fs *FileSystem) ReadInode(ino Ino) (Inode, error) {
	offset, inodeSize := fs.LocateInode(ino)
	inodeBuf := make([]byte, inodeSize)
	if err := fs.Volume.Read(offset, inodeBuf); err != nil {
		return Inode{}, fmt.Errorf("reading inode at `%#x`: %w", ino, err)
	}
	inode, err := DecodeInode(
		ino,
		fs.Superblock.RevLevel,
		(*[InodeBufferSize]byte)(inodeBuf),
	)
	if err != nil {
		return Inode{}, fmt.Errorf("reading inode at `%#x`: %w", ino, err)
	}
	return inode, nil
}

func (fs *FileSystem) RefitInodeCache() error {
	for len(fs.InodeCache) > 10 {
		flushed := false
		for {
			usedIno, ok := fs.CacheQueue.PopFront()
			if !ok {
				break
			}

			if _, exists := fs.ReusedInos[usedIno]; exists {
				delete(fs.ReusedInos, usedIno)
				fs.CacheQueue.PushBack(usedIno)
			} else {
				if err := fs.FlushIno(usedIno); err != nil {
					return fmt.Errorf("refitting inode cache: %w", err)
				}
				flushed = true
				break
			}
		}

		if !flushed {
			// roughly translating these lines:
			// let random_ino = *fs.inode_cache.iter().next().unwrap().0;
			// try!(flush_ino(fs, random_ino));
			for ino := range fs.InodeCache {
				if err := fs.FlushIno(ino); err != nil {
					return fmt.Errorf("refitting inode cache: %w", err)
				}
				break
			}
		}
	}

	return nil
}

func (fs *FileSystem) FlushIno(ino Ino) error {
	if inode, exists := fs.InodeCache[ino]; exists {
		delete(fs.InodeCache, ino)
		delete(fs.ReusedInos, ino)
		if _, exists := fs.DirtyInos[ino]; exists {
			delete(fs.DirtyInos, ino)
			if err := fs.WriteInode(&inode); err != nil {
				return fmt.Errorf("flushing ino `%#x`: %w", ino, err)
			}
			return nil
		}
	}
	return nil
}

func (fs *FileSystem) WriteInode(inode *Inode) error {
	offset, inodeSize := fs.LocateInode(inode.Ino)
	inodeBuf := make([]byte, inodeSize)
	if err := fs.Volume.Read(offset, inodeBuf); err != nil {
		return fmt.Errorf("writing inode `%#x`: %w", inode.Ino, err)
	}
	if err := inode.Encode(
		fs.Superblock.RevLevel,
		(*[InodeBufferSize]byte)(inodeBuf),
	); err != nil {
		return fmt.Errorf("writing inode `%#x`: %w", inode.Ino, err)
	}
	if err := fs.Volume.Write(offset, inodeBuf); err != nil {
		return fmt.Errorf("writing inode `%#x`: %w", inode.Ino, err)
	}
	return nil
}

func (fs *FileSystem) LocateInode(ino Ino) (uint64, uint64) {
	groupID, localID := fs.GetInoGroup(ino)
	inodeSize := uint64(fs.Superblock.InodeSize)
	inodeTable := uint64(fs.Groups[groupID].Desc.InodeTable)
	offset := inodeTable*fs.BlockSize() + localID*inodeSize
	return offset, inodeSize
}

func (fs *FileSystem) GetInoGroup(ino Ino) (GroupID, uint64) {
	groupSize := GroupID(fs.Superblock.InodesPerGroup)
	return GroupID(ino-1) / groupSize, uint64(ino-1) % uint64(groupSize)
}

func (fs *FileSystem) OpenFile(ino Ino) (FileHandle, error) {
	inode, err := fs.GetInode(ino)
	if err != nil {
		return FileHandle{}, fmt.Errorf("opening file: %w", err)
	}
	if inode.Mode.FileType == FileTypeRegular {
		return FileHandle{ino}, nil
	}

	return FileHandle{}, fmt.Errorf(
		"opening ino `%#x` as regular file: %w",
		ino,
		ErrInvalidFileType{
			Wanted: FileTypeRegular,
			Found:  inode.Mode.FileType,
		},
	)
}

func (fs *FileSystem) ReadFile(
	handle *FileHandle,
	offset uint64,
	b []byte,
) (uint64, error) {
	inode, err := fs.GetInode(handle.ino)
	if err != nil {
		return 0, fmt.Errorf("reading file: %w", handle.ino, err)
	}
	n, err := fs.ReadInodeData(&inode, offset, b)
	if err != nil {
		return n, fmt.Errorf("reading inode data: %w", handle.ino, err)
	}
	return n, nil
}

func (fs *FileSystem) ReadInodeData(
	inode *Inode,
	offset uint64,
	b []byte,
) (uint64, error) {
	blockSize := fs.BlockSize()
	maxLength := min(uint64(len(b)), inode.Size-offset)
	var chunkBegin uint64
	for chunkBegin < maxLength {
		chunkBlock := (offset + chunkBegin) / blockSize
		chunkOffset := (offset + chunkBegin) % blockSize
		chunkLength := min(maxLength-chunkBegin, blockSize-chunkOffset)
		if err := fs.ReadInodeBlock(
			inode,
			chunkBlock,
			chunkOffset,
			b[chunkBegin:chunkLength],
		); err != nil {
			return chunkBegin, fmt.Errorf("reading inode data: %w", err)
		}
		chunkBegin += chunkLength
	}
	return chunkBegin, nil
}

func (fs *FileSystem) ReadInodeBlock(
	inode *Inode,
	inodeBlock uint64,
	offset uint64,
	b []byte,
) error {
	blockSize := fs.BlockSize()
	if offset+uint64(len(b)) > blockSize {
		panic(fmt.Sprintf(
			"offset `%d` + buffer length `%d` must be less than block size "+
				"`%d`",
			offset,
			len(b),
			blockSize,
		))
	}

	realBlock, ok, err := fs.GetInodeBlock(inode, inodeBlock)
	if err != nil {
		return fmt.Errorf(
			"reading block for inode at offset `%#x`: %w",
			offset,
			err,
		)
	}
	if !ok {
		return fmt.Errorf(
			"reading block for inode at offset `%#x`: %w",
			offset,
			ErrBlockOutOfRange{inodeBlock},
		)
	}

	blockOffset := realBlock*blockSize + offset
	if err := fs.Volume.Read(blockOffset, b); err != nil {
		return fmt.Errorf(
			"reading block for inode `%#x` at block `%#x` and offset `%#x`: "+
				"%w",
			inode.Ino,
			inodeBlock,
			offset,
			err,
		)
	}
	return nil
}

func (fs *FileSystem) GetInodeBlock(
	inode *Inode,
	inodeBlock uint64,
) (uint64, bool, error) {
	pos := fs.InodeBlockToPos(inodeBlock)
	switch pos.Level {
	case PosLevel0:
		block0 := uint64(inode.Block[pos.Data[0]])
		if block0 == 0 {
			return 0, false, nil
		}
		return block0, true, nil
	case PosLevel1:
		block1 := uint64(inode.Block[12])
		if block1 == 0 {
			return 0, false, nil
		}
		block0, err := fs.ReadIndirect(block1, pos.Data[0])
		if err != nil {
			return 0, false, fmt.Errorf(
				"getting block `%#x` for inode `%#x`: %w",
				inodeBlock,
				inode.Ino,
				err,
			)
		}
		if block0 == 0 {
			return 0, false, nil
		}
		return block0, true, nil
	case PosLevel2:
		level1, level0 := pos.Data[0], pos.Data[1]
		block2 := uint64(inode.Block[13])
		if block2 == 0 {
			return 0, false, nil
		}
		block1, err := fs.ReadIndirect(block2, level1)
		if err != nil {
			return 0, false, fmt.Errorf(
				"getting block `%#x` for inode `%#x`: %w",
				inodeBlock,
				inode.Ino,
				err,
			)
		}
		if block1 == 0 {
			return 0, false, nil
		}
		block0, err := fs.ReadIndirect(block1, level0)
		if err != nil {
			return 0, false, fmt.Errorf(
				"getting block `%#x` for inode `%#x`: %w",
				inodeBlock,
				inode.Ino,
				err,
			)
		}
		if block0 == 0 {
			return 0, false, nil
		}
		return block0, true, nil
	case PosLevel3:
		level2, level1, level0 := pos.Data[0], pos.Data[1], pos.Data[2]
		block3 := uint64(inode.Block[14])
		if block3 == 0 {
			return 0, false, nil
		}
		block2, err := fs.ReadIndirect(block3, level2)
		if err != nil {
			return 0, false, fmt.Errorf(
				"getting block `%#x` for inode `%#x`: %w",
				inodeBlock,
				inode.Ino,
				err,
			)
		}
		if block2 == 0 {
			return 0, false, nil
		}
		block1, err := fs.ReadIndirect(block2, level1)
		if err != nil {
			return 0, false, fmt.Errorf(
				"getting block `%#x` for inode `%#x`: %w",
				inodeBlock,
				inode.Ino,
				err,
			)
		}
		if block1 == 0 {
			return 0, false, nil
		}
		block0, err := fs.ReadIndirect(block1, level0)
		if err != nil {
			return 0, false, fmt.Errorf(
				"getting block `%#x` for inode `%#x`: %w",
				inodeBlock,
				inode.Ino,
				err,
			)
		}
		if block0 == 0 {
			return 0, false, nil
		}
		return block0, true, nil
	case PosOutOfRange:
		return 0, false, fmt.Errorf(
			"getting block `%#x` for inode `%#x`: %w",
			inodeBlock,
			inode.Ino,
			ErrBlockOutOfRange{inodeBlock},
		)
	default:
		panic(fmt.Sprint("invalid BlockPosLevel: %d", pos.Level))
	}
}

func (fs *FileSystem) ReadIndirect(
	indirectBlock uint64,
	entry uint64,
) (uint64, error) {
	var b [4]byte
	blockSize := fs.BlockSize()
	entryOffset := indirectBlock*blockSize + entry*4
	if entry >= blockSize/4 {
		panic(fmt.Sprintf(
			"entry `%d` should be less than a quarter of the block size `%d`",
			entry,
			blockSize/4,
		))
	}
	if err := fs.Volume.Read(entryOffset, b[:]); err != nil {
		return 0, fmt.Errorf(
			"reading indirect block `%#x` at entry `%#x`: %w",
			indirectBlock,
			entry,
			err,
		)
	}
	return uint64(DecodeUint32(b[0], b[1], b[2], b[3])), nil
}

func (fs *FileSystem) WriteFile(
	handle *FileHandle,
	offset uint64,
	b []byte,
) (uint64, error) {
	inode, err := fs.GetInode(handle.ino)
	if err != nil {
		return 0, fmt.Errorf("writing file: %w", err)
	}

	n, err := fs.WriteInodeData(&inode, offset, b)
	if err != nil {
		return n, fmt.Errorf("writing file: %w", err)
	}

	return n, nil
}

func (fs *FileSystem) WriteInodeData(
	inode *Inode,
	offset uint64,
	b []byte,
) (uint64, error) {
	blockSize := fs.BlockSize()
	var chunkBegin uint64
	for chunkBegin < uint64(len(b)) {
		chunkBlock := (offset + chunkBegin) / blockSize
		chunkOffset := (offset + chunkBegin) % blockSize
		chunkLength := min(uint64(len(b))-chunkBegin, blockSize-chunkOffset)
		if err := fs.WriteInodeBlock(
			inode,
			chunkBlock,
			chunkOffset,
			b[chunkBegin:chunkLength],
		); err != nil {
			return chunkBegin, fmt.Errorf("writing inode data: %w", err)
		}
		chunkBegin += chunkLength
	}

	if minSize := offset + chunkBegin; inode.Size < minSize {
		inode.Size = minSize
		if err := fs.UpdateInode(inode); err != nil {
			return chunkBegin, fmt.Errorf("writing inode data: %w", err)
		}
	}

	return chunkBegin, nil
}

func (fs *FileSystem) WriteInodeBlock(
	inode *Inode,
	inodeBlock uint64,
	offset uint64,
	b []byte,
) error {
	blockSize := fs.BlockSize()
	if uint64(len(b))+offset > blockSize {
		panic(fmt.Sprintf(
			"offset `%d` + len(buffer) `%d` exceeds block size `%d`",
			offset,
			len(b),
			blockSize,
		))
	}
	realBlock, ok, err := fs.GetInodeBlock(inode, inodeBlock)
	if err != nil {
		return fmt.Errorf(
			"writing block for inode at offset `%#x`: %w",
			offset,
			err,
		)
	}
	if !ok {
		block, err := fs.AllocInodeBlock(inode)
		if err != nil {
			return fmt.Errorf(
				"writing block `%#x` for inode at offset `%#x`: %w",
				inodeBlock,
				offset,
				err,
			)
		}
		realBlock = block
	}

	blockOffset := realBlock*blockSize + offset
	if err := fs.Volume.Write(blockOffset, b); err != nil {
		return fmt.Errorf(
			"writing block `%#x` for inode `%#x` at offset `%#x`: %w",
			inodeBlock,
			inode.Ino,
			offset,
			err,
		)
	}

	return nil
}

func (fs *FileSystem) AllocInodeBlock(inode *Inode) (uint64, error) {
	inodeGroupID, _ := fs.GetInoGroup(inode.Ino)
	block, ok, err := fs.AllocBlock(inodeGroupID)
	if err != nil {
		return 0, fmt.Errorf(
			"allocating block for inode `%#x`: %w",
			inode.Ino,
			err,
		)
	}
	if !ok {
		return 0, fmt.Errorf(
			"allocating block for inode `%#x`: %w",
			inode.Ino,
			NoFreeBlocksErr,
		)
	}
	inode.Size512 += uint32(fs.BlockSize() / 512)
	if err := fs.UpdateInode(inode); err != nil {
		return 0, fmt.Errorf(
			"allocating block for inode `%#x`: %w",
			inode.Ino,
			err,
		)
	}
	return block, nil
}

//	pub fn update_inode(fs: &mut Filesystem, inode: &Inode) -> Result<()> {
//	  use std::collections::hash_map::Entry;
//
//	  fs.dirty_inos.insert(inode.ino);
//	  match fs.inode_cache.entry(inode.ino) {
//	    Entry::Occupied(mut occupied) => {
//	      occupied.insert(inode.clone());
//	      fs.reused_inos.insert(inode.ino);
//	      return Ok(())
//	    },
//	    Entry::Vacant(vacant) => {
//	      vacant.insert(inode.clone());
//	      fs.cache_queue.push_back(inode.ino);
//	    },
//	  }
//
//	  refit_inode_cache(fs)
//	}
func (fs *FileSystem) UpdateInode(inode *Inode) error {
	fs.DirtyInos[inode.Ino] = struct{}{}
	if _, exists := fs.InodeCache[inode.Ino]; exists {
		fs.InodeCache[inode.Ino] = *inode
		fs.ReusedInos[inode.Ino] = struct{}{}
		return nil
	}
	fs.InodeCache[inode.Ino] = *inode
	fs.CacheQueue.PushBack(inode.Ino)
	if err := fs.RefitInodeCache(); err != nil {
		return fmt.Errorf("updating inode `%#x`: %w", inode.Ino, err)
	}
	return nil
}

func (fs *FileSystem) AllocBlock(firstGroupID GroupID) (uint64, bool, error) {
	return fs.Alloc(firstGroupID, (*FileSystem).AllocBlockInGroup)
}

//	fn alloc_block_in_group(fs: &mut Filesystem, group_idx: u64) -> Result<Option<u64>> {
//	  let group_id = group_idx as usize;
//	  if fs.groups[group_id].desc.free_blocks_count == 0 {
//	    return Ok(None)
//	  }
//
//	  match find_zero_bit_in_bitmap(&fs.groups[group_id].block_bitmap[..]) {
//	    Some((byte, bit)) => {
//	  	fs.groups[group_id].block_bitmap[byte as usize] |= 1 << bit;
//	  	fs.groups[group_id].desc.free_blocks_count -= 1;
//	  	fs.groups[group_id].dirty = true;
//	  	fs.superblock.free_blocks_count -= 1;
//	  	fs.superblock_dirty = true;
//	  	Ok(Some(group_idx * fs.superblock.blocks_per_group as u64 +
//	  			fs.superblock.first_data_block as u64 +
//	  			byte * 8 + bit))
//	    },
//	    None => Ok(None),
//	  }
//	}
func (fs *FileSystem) AllocBlockInGroup(groupID GroupID) (uint64, bool, error) {
	if fs.Groups[groupID].Desc.FreeBlocksCount == 0 {
		return 0, false, nil
	}

	byt, bit, ok := fs.Groups[groupID].BlockBitmap.FindZeroBit()
	if !ok {
		return 0, false, nil
	}

	fs.Groups[groupID].BlockBitmap[byt] |= 1 << bit
	fs.Groups[groupID].Desc.FreeBlocksCount--
	fs.Groups[groupID].Dirty = true
	fs.Superblock.FreeBlocksCount--
	fs.SuperblockDirty = true
	return uint64(groupID)*uint64(fs.Superblock.BlocksPerGroup) +
		uint64(fs.Superblock.FirstDataBlock) +
		byt*8 + bit, true, nil
}

// fn alloc(fs: &mut Filesystem, first_group_idx: u64,
//
//	alloc_in_group: fn(&mut Filesystem, u64) -> Result<Option<u64>>)
//	-> Result<Option<u64>>
//
//	{
//	  Ok(match try!(alloc_in_group(fs, first_group_idx)) {
//	    Some(resource) => Some(resource),
//	    None => {
//	      let group_count = fs.group_count();
//	      for group_idx in (first_group_idx..group_count).chain(0..first_group_idx) {
//	        if let Some(resource) = try!(alloc_in_group(fs, group_idx)) {
//	          return Ok(Some(resource));
//	        }
//	      }
//	      None
//	    }
//	  })
//	}
func (fs *FileSystem) Alloc(
	firstGroupID GroupID,
	allocInGroup func(*FileSystem, GroupID) (uint64, bool, error),
) (uint64, bool, error) {
	resource, ok, err := allocInGroup(fs, firstGroupID)
	if err != nil {
		return resource, ok, err
	}
	if ok {
		return resource, true, nil
	}
	groupCount := GroupID(fs.GroupCount())
	for _, rng := range [2][2]GroupID{
		{firstGroupID, groupCount},
		{0, firstGroupID},
	} {
		for groupID := rng[0]; groupID < rng[1]; groupID++ {
			resource, ok, err := allocInGroup(fs, groupID)
			if err != nil {
				return resource, ok, err
			}
			if ok {
				return resource, true, nil
			}
		}
	}

	return 0, false, nil
}

func (fs *FileSystem) AllocTables() error {
	for i := range fs.Groups {
		if err := fs.AllocGroupTable(GroupID(i)); err != nil {
			return err
		}
	}

	return nil
}

func (fs *FileSystem) AllocGroupTable(group GroupID) error {
	var (
		groupDesc = fs.Groups[group].Desc
		err       error
	)
	// TODO: Do we need to handle the possibility that a groupdescriptor could
	// be split across groups? I.e., if a call to allocBlockInGroup returns
	// NoFreeBlocksErr, do we need to increment the `group` variable and retry?

	// TODO: The last group may not have the full number of blocks
	groupDesc.FreeBlocksCount = uint16(fs.Superblock.BlocksPerGroup)

	// TODO: Before allocating the block an inode bitmaps, we need to allocate
	// blocks for writing the group descriptor table itself

	// Allocate block and inode bitmaps
	groupDesc.BlockBitmap, err = fs.allocBlockInGroup(group)
	if err != nil {
		return fmt.Errorf(
			"allocating table for group `%#x`: allocating block bitmap: %w",
			group,
			err,
		)
	}
	groupDesc.FreeBlocksCount--

	groupDesc.InodeBitmap, err = fs.allocBlockInGroup(group)
	if err != nil {
		return fmt.Errorf(
			"allocating table for group `%#x`: allocating inode bitmap: %w",
			group,
			err,
		)
	}
	groupDesc.FreeBlocksCount--

	inodeBlocks, err := fs.AllocateInodeTable(group)
	if err != nil {
		return fmt.Errorf("allocating table for group `%#x`: %w", group, err)
	}
	groupDesc.FreeBlocksCount -= inodeBlocks

	return nil
}

func (fs *FileSystem) AllocateInodeTable(group GroupID) (uint16, error) {
	inodeBlocks := fs.Superblock.InodesPerGroup *
		uint32(fs.Superblock.InodeSize) /
		uint32(fs.BlockSize())

	// allocate the first block of the inode table; note the block in the
	// GroupDesc.InodeTable field.
	block, err := fs.allocBlockInGroup(group)
	if err != nil {
		return 0, fmt.Errorf("allocating block 0 of inode table: %w", err)
	}
	fs.Groups[group].Desc.InodeTable = block

	// allocate the remaining inode table blocks
	for i := uint32(1); i < inodeBlocks; i++ {
		if _, err := fs.allocBlockInGroup(group); err != nil {
			return i + 1, fmt.Errorf(
				"allocating block %d of inode table: %w",
				i,
				err,
			)
		}
	}

	return inodeBlocks, nil
}

func (fs *FileSystem) allocBlockInGroup(group GroupID) (uint32, error) {
	block, ok, err := fs.Alloc(group, (*FileSystem).AllocBlockInGroup)
	if err != nil {
		return uint32(block), err
	}
	if !ok {
		return uint32(block), NoFreeBlocksErr
	}
	return uint32(block), nil
}

///*
// * Return the first block (inclusive) in a group
// */
// blk64_t ext2fs_group_first_block2(ext2_filsys fs, dgrp_t group)
// {
//	 return fs->super->s_first_data_block +
//		 EXT2_GROUPS_TO_BLOCKS(fs->super, group);
// }

func (fs *FileSystem) GroupLastBlock(group GroupID) uint64 {
	if group == GroupID(len(fs.Groups)-1) {
		return uint64(fs.Superblock.BlocksCount) - 1
	}
	return fs.GroupFirstBlock(group) + uint64(fs.Superblock.BlocksPerGroup) - 1
}

// /*
//  * Return the last block (inclusive) in a group
//  */
// blk64_t ext2fs_group_last_block2(ext2_filsys fs, dgrp_t group)
// {
// 	return (group == fs->group_desc_count - 1 ?
// 		ext2fs_blocks_count(fs->super) - 1 :
// 		ext2fs_group_first_block2(fs, group) +
// 			(fs->super->s_blocks_per_group - 1));
// }

func (fs *FileSystem) CloseFile(handle FileHandle) error {
	if err := fs.FlushIno(handle.ino); err != nil {
		return fmt.Errorf("closing file: %w", err)
	}
	return nil
}

func (fs *FileSystem) InodeBlockToPos(inodeBlock uint64) BlockPos {
	if inodeBlock < 12 {
		return BlockPosLevel0(inodeBlock)
	}

	indirect1Size := fs.BlockSize() / 4
	if inodeBlock < 12+indirect1Size {
		return BlockPosLevel1(inodeBlock - 12)
	}

	indirect2Size := indirect1Size * indirect1Size
	if inodeBlock < 12+indirect1Size+indirect2Size {
		base := inodeBlock - 12 - indirect1Size
		return BlockPosLevel2(base/indirect1Size, base%indirect1Size)
	}

	indirect3Size := indirect1Size * indirect2Size
	if inodeBlock < 12+indirect1Size+indirect2Size+indirect3Size {
		base := inodeBlock - 12 - indirect1Size - indirect2Size
		return BlockPosLevel3(
			base/indirect2Size,
			(base%indirect2Size)/indirect1Size,
			(base%indirect2Size)%indirect1Size,
		)
	}

	return BlockPosOutOfRange()
}

//	pub fn flush_fs(fs: &mut Filesystem) -> Result<()> {
//	  let dirty_inos = fs.dirty_inos.clone();
//	  for dirty_ino in dirty_inos {
//	    try!(flush_ino(fs, dirty_ino));
//	  }
//
//	  for group_idx in 0..fs.group_count() {
//	    try!(flush_group(fs, group_idx));
//	  }
//
//	  flush_superblock(fs, true)
//	}
func (fs *FileSystem) Flush() error {
	for ino := range fs.DirtyInos {
		if err := fs.FlushIno(ino); err != nil {
			return fmt.Errorf("flushing filesystem: %w", err)
		}
	}

	groupCount := GroupID(fs.GroupCount())
	for groupID := GroupID(0); groupID < groupCount; groupID++ {
		if err := fs.FlushGroup(groupID); err != nil {
			return fmt.Errorf("flushing filesystem: %w", err)
		}
	}

	if err := fs.FlushSuperblock(true); err != nil {
		return fmt.Errorf("flushing filesystem: %w", err)
	}

	return nil
}

//	pub fn flush_group(fs: &mut Filesystem, group_idx: u64) -> Result<()> {
//	  if fs.groups[group_idx as usize].dirty {
//	    try!(write_group(fs, group_idx));
//	    fs.groups[group_idx as usize].dirty = false;
//	  }
//	  Ok(())
//	}
func (fs *FileSystem) FlushGroup(groupID GroupID) error {
	if fs.Groups[groupID].Dirty {
		if err := fs.WriteGroup(groupID); err != nil {
			return fmt.Errorf("flushing group `%#x`: %w", groupID, err)
		}
		fs.Groups[groupID].Dirty = false
	}
	return nil
}

//	fn write_group(fs: &mut Filesystem, group_idx: u64) -> Result<()> {
//	  let group_desc = fs.groups[group_idx as usize].desc;
//	  let table_block = fs.superblock.first_data_block as u64 + 1;
//	  try!(write_group_desc(fs, table_block, group_idx, &group_desc));
//
//	  let block_bitmap_offset = group_desc.block_bitmap as u64 * fs.block_size();
//	  try!(fs.volume.write(block_bitmap_offset,
//	    &fs.groups[group_idx as usize].block_bitmap[..]));
//
//	  let inode_bitmap_offset = group_desc.inode_bitmap as u64 * fs.block_size();
//	  try!(fs.volume.write(inode_bitmap_offset,
//	    &fs.groups[group_idx as usize].inode_bitmap[..]));
//
//	  Ok(())
//	}
func (fs *FileSystem) WriteGroup(groupID GroupID) error {
	groupDesc := fs.Groups[groupID].Desc
	tableBlock := uint64(fs.Superblock.FirstDataBlock) + 1
	if err := fs.WriteGroupDesc(tableBlock, groupID, &groupDesc); err != nil {
		return fmt.Errorf("writing group `%#x`: %w", groupID, err)
	}

	blockSize := fs.BlockSize()
	blockBitmapOffset := uint64(groupDesc.BlockBitmap) * blockSize
	if err := fs.Volume.Write(
		blockBitmapOffset,
		[]byte(fs.Groups[groupID].BlockBitmap),
	); err != nil {
		return fmt.Errorf(
			"writing group `%#x`: writing block bitmap: %w",
			groupID,
			err,
		)
	}

	inodeBitmapOffset := uint64(groupDesc.InodeBitmap) * blockSize
	if err := fs.Volume.Write(
		inodeBitmapOffset,
		[]byte(fs.Groups[groupID].InodeBitmap),
	); err != nil {
		return fmt.Errorf(
			"writing group `%#x`: writing inode bitmap: %w",
			groupID,
			err,
		)
	}

	return nil
}

// fn write_group_desc(fs: &mut Filesystem, table_block: u64, group_idx: u64,
//
//	desc: &GroupDesc) -> Result<()> {
//		  let offset = table_block * fs.block_size() + group_idx * 32;
//		  let mut desc_buf = make_buffer(32);
//		  try!(fs.volume.read(offset, &mut desc_buf[..]));
//		  try!(encode_group_desc(&fs.superblock, desc, &mut desc_buf[..]));
//		  fs.volume.write(offset, &desc_buf[..])
//		}
func (fs *FileSystem) WriteGroupDesc(
	tableBlock uint64,
	groupID GroupID,
	desc *GroupDesc,
) error {
	offset := tableBlock*fs.BlockSize() + uint64(groupID)*GroupDescSize
	var descBuf [GroupDescSize]byte
	if err := fs.Volume.Read(offset, descBuf[:]); err != nil {
		return fmt.Errorf(
			"writing desc for group `%#x` at table block `%#x`: %w",
			groupID,
			tableBlock,
			err,
		)
	}
	desc.Encode(&descBuf)
	if err := fs.Volume.Write(offset, descBuf[:]); err != nil {
		return fmt.Errorf(
			"writing desc for group `%#x` at table block `%#x`: %w",
			groupID,
			tableBlock,
			err,
		)
	}

	return nil
}

// errcode_t ext2fs_new_block3(ext2_filsys fs, blk64_t goal,
//
//	ext2fs_block_bitmap map, blk64_t *ret,
//	struct blk_alloc_ctx *ctx)
//
//	{
//		errcode_t retval;
//		blk64_t	b = 0;
//		errcode_t (*gab)(ext2_filsys fs, blk64_t goal, blk64_t *ret);
//		errcode_t (*gab2)(ext2_filsys, blk64_t, blk64_t *,
//				  struct blk_alloc_ctx *);
//
//		EXT2_CHECK_MAGIC(fs, EXT2_ET_MAGIC_EXT2FS_FILSYS);
//
//		if (!map) {
//			/*
//			 * In case there are clients out there whose get_alloc_block
//			 * handlers call ext2fs_new_block2 with a NULL block map,
//			 * temporarily swap out the function pointer so that we don't
//			 * end up in an infinite loop.
//			 */
//			 if (fs->get_alloc_block2) {
//				gab2 = fs->get_alloc_block2;
//				fs->get_alloc_block2 = NULL;
//				retval = gab2(fs, goal, &b, ctx);
//				fs->get_alloc_block2 = gab2;
//				goto allocated;
//			} else if (fs->get_alloc_block) {
//				gab = fs->get_alloc_block;
//				fs->get_alloc_block = NULL;
//				retval = gab(fs, goal, &b);
//				fs->get_alloc_block = gab;
//				goto allocated;
//			}
//		}
//		if (!map)
//			map = fs->block_map;
//		if (!map)
//			return EXT2_ET_NO_BLOCK_BITMAP;
//		if (!goal || (goal >= ext2fs_blocks_count(fs->super)))
//			goal = fs->super->s_first_data_block;
//		goal &= ~EXT2FS_CLUSTER_MASK(fs);
//
//		retval = ext2fs_find_first_zero_block_bitmap2(map,
//				goal, ext2fs_blocks_count(fs->super) - 1, &b);
//		if ((retval == ENOENT) && (goal != fs->super->s_first_data_block))
//			retval = ext2fs_find_first_zero_block_bitmap2(map,
//				fs->super->s_first_data_block, goal - 1, &b);
//
// allocated:
//
//		if (retval == ENOENT)
//			return EXT2_ET_BLOCK_ALLOC_FAIL;
//		if (retval)
//			return retval;
//
//		ext2fs_clear_block_uninit(fs, ext2fs_group_of_blk2(fs, b));
//		*ret = b;
//		return 0;
//	}
// func (fs *FileSystem) NewBlock() error {}

type ErrInvalidFileType struct {
	Wanted, Found FileType
}

func (err ErrInvalidFileType) Error() string {
	return fmt.Sprintf(
		"invalid file type: wanted `%s`; found `%s`",
		err.Wanted,
		err.Found,
	)
}

type ErrBlockOutOfRange struct {
	Block uint64
}

func (err ErrBlockOutOfRange) Error() string {
	return fmt.Sprintf("block `%#x` is out of range", err.Block)
}

var NoFreeBlocksErr = errors.New("no free blocks remain for files")
