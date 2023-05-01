package fs

import (
	"fmt"
	"io"
	"time"
)

type FileSystem struct {
	Volume     io.ReadWriteSeeker
	Superblock Superblock
	Descriptor Descriptor
	InodeCache Cache
	DirtyInos  InoSet
}

type FileSystemParams struct {
	Volume        io.ReadWriteSeeker
	BlockSize     Byte
	Blocks        Block
	Inodes        Ino
	CacheCapacity int
}

func NewFileSystem(params *FileSystemParams) FileSystem {
	return FileSystem{
		Volume: params.Volume,
		Superblock: NewSuperblock(
			params.BlockSize,
			params.Blocks,
			params.Inodes,
		),
		// TODO: before returning; we should mark the reserved blocks and
		// inodes.
		Descriptor: NewDescriptor(params.Blocks, params.Inodes),
		InodeCache: NewCache(params.CacheCapacity),
		DirtyInos:  NewInoSet(),
	}
}

func LoadFileSystem(
	volume io.ReadWriteSeeker,
	cacheCapacity int,
) (FileSystem, error) {
	fs := FileSystem{
		Volume:     volume,
		InodeCache: NewCache(cacheCapacity),
		DirtyInos:  NewInoSet(),
	}
	if err := ReadSuperblock(&fs); err != nil {
		return FileSystem{}, fmt.Errorf("loading filesystem: %w", err)
	}
	fs.Descriptor = NewDescriptor(
		fs.Superblock.BlockCount,
		fs.Superblock.InodeCount,
	)
	return fs, nil
}

func (fs *FileSystem) Init() error {
	root, err := GetInode(fs, InoRoot)
	if err != nil {
		return fmt.Errorf(
			"initializing file system: allocating root inode: %w",
			err,
		)
	}
	now := timestamp(time.Now())
	root.Mode.Type = FileTypeDir
	root.Mode.AccessRights = 0744
	root.Attr.ATime = now
	root.Attr.CTime = now
	root.Attr.MTime = now
	root.Size = fs.Superblock.BlockSize
	root.LinksCount = 2
	if err := UpdateInode(fs, &root); err != nil {
		return fmt.Errorf("initializing file system: %w", err)
	}
	return nil
}

func InitFileSystem(params *FileSystemParams) error {
	fs := NewFileSystem(params)
	if err := fs.Init(); err != nil {
		return err
	}
	if err := fs.Flush(); err != nil {
		return fmt.Errorf("initializing file system: %w", err)
	}
	return nil
}

func timestamp(t time.Time) uint32 {
	// take the most significant 32 bits
	return uint32(t.Unix() >> 32)
}

func (fs *FileSystem) Flush() error {
	// TODO: only write the descriptor if it's dirty
	if err := WriteSuperblock(fs); err != nil {
		return fmt.Errorf("flushing filesystem: %w", err)
	}
	// TODO: only write the descriptor if it's dirty
	if err := WriteDescriptor(fs); err != nil {
		return fmt.Errorf("flushing filesystem: %w", err)
	}
	for ino := range fs.DirtyInos {
		if err := FlushIno(fs, ino); err != nil {
			return fmt.Errorf("flushing filesystem: %w", err)
		}
	}
	return nil
}
