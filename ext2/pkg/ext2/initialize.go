package ext2

func NewFileSystem(
	superblock *Superblock,
	volume Volume,
) *FileSystem {
	fs := &FileSystem{
		Volume:          volume,
		Superblock:      *superblock,
		SuperblockBytes: &[SuperblockSize]byte{},
		SuperblockDirty: true,
		Groups:          nil,
		InodeCache:      map[Ino]Inode{},
		DirtyInos:       map[Ino]struct{}{},
		ReusedInos:      map[Ino]struct{}{},
		CacheQueue:      NewRing(),
		BlockBitmap:     make(DynamicBitmap, superblock.BlocksCount/8), // one bit per block
	}

	fs.Groups = make([]Group, fs.GroupCount())
	for groupID := range fs.Groups {
		blockForGroup := SuperblockOffset + uint32(groupID)*superblock.BlocksPerGroup
		fs.Groups[groupID] = Group{
			Idx: GroupID(groupID),
			Desc: GroupDesc{
				BlockBitmap:     blockForGroup,
				InodeBitmap:     blockForGroup + 1,
				InodeTable:      0,
				FreeBlocksCount: 0,
				FreeInodesCount: 0,
				UsedDirsCount:   0,
			},
			BlockBitmap: make(DynamicBitmap, 1024<<superblock.LogBlockSize),
			InodeBitmap: make(DynamicBitmap, 0),
			Dirty:       true,
		}
	}
}
