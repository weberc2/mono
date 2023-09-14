package main

const RootIno uint64 = 2

type FileSystem struct {
	Volume          Volume
	Superblock      Superblock
	SuperblockBytes *[1024]byte
	SuperblockDirty bool
	Groups          []Group
	InodeCache      map[uint64]Inode
	DirtyInos       map[uint64]struct{}
	ReusedInos      map[uint64]struct{}
	CacheQueue      *Ring
}

func (fs *FileSystem) BlockSize() uint64 {
	return 1024 << fs.Superblock.LogBlockSize
}

func (fs *FileSystem) GroupCount() uint64 {
	a := uint64(fs.Superblock.BlocksCount)
	b := uint64(fs.Superblock.BlocksPerGroup)
	return (a + b - 1) / b
}

// func (fs *FileSystem) Mount(volume Volume) error {
// 	var superblockBytes [1024]byte
// 	if err := volume.Read(1024, superblockBytes[:]); err != nil {
// 		return fmt.Errorf("mounting filesystem: %w", err)
// 	}
//
// 	sb, err := DecodeSuperblock(&superblockBytes, false)
// 	if err != nil {
// 		return fmt.Errorf("mounting filesystem: %w", err)
// 	}
//
// 	tmp := FileSystem{
// 		Volume:          volume,
// 		Superblock:      sb,
// 		SuperblockBytes: &superblockBytes,
// 		SuperblockDirty: false,
// 		Groups:          nil,
// 		InodeCache:      map[uint64]Inode{},
// 		DirtyInos:       map[uint64]struct{}{},
// 		ReusedInos:      map[uint64]struct{}{},
// 		CacheQueue:      nil, // empty ring
// 	}
//
// 	tmp.Groups = make([]Group, tmp.GroupCount())
// 	for i := uint64(0); i < tmp.GroupCount(); i++ {
// 		tmp.Groups[i] = ReadGroup(&tmp, i)
// 	}
// }
//
