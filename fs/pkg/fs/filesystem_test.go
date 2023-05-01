package fs

import (
	"io"
	"testing"
)

func TestFileSystem(t *testing.T) {
	volume := NewBuffer(make([]byte, 1024*1024))

	if err := InitFileSystem(&FileSystemParams{
		Volume:        volume,
		BlockSize:     DefaultBlockSize,
		Blocks:        1024 * 1024,
		Inodes:        50,
		CacheCapacity: 10,
	}); err != nil {
		t.Fatalf("InitFileSystem(): unexpected err: %v", err)
	}

	if _, err := volume.Seek(0, io.SeekStart); err != nil {
		t.Fatalf("Buffer.Seek(): unexpected err: %v", err)
	}

	fs, err := LoadFileSystem(volume, 10)
	if err != nil {
		t.Fatalf("LoadFileSystem(): unexpected err: %v", err)
	}

	// Confirm superblock matches expectations
	if fs.Superblock.BlockSize != DefaultBlockSize {
		t.Fatalf(
			"Superblock.BlockSize: wanted `%d`; found `%d`",
			DefaultBlockSize,
			fs.Superblock.BlockSize,
		)
	}
	if fs.Superblock.BlockCount != 1024*1024 {
		t.Fatalf(
			"Superblock.BlockCount: wanted `%d`; found `%d`",
			1024*1024,
			fs.Superblock.BlockCount,
		)
	}
	if fs.Superblock.InodeCount != 50 {
		t.Fatalf(
			"Superblock.InodeCount: wanted `50`; found `%d`",
			fs.Superblock.InodeCount,
		)
	}

	// Confirm descriptor matches expectations
	wantedDescriptor := NewDescriptor(
		fs.Superblock.BlockCount,
		fs.Superblock.InodeCount,
	)
	if !wantedDescriptor.Equal(&fs.Descriptor) {
		t.Fatalf(
			"Descriptor: wanted `%s`; found `%s`",
			wantedDescriptor.Debug(),
			fs.Descriptor.Debug(),
		)
	}

	// Confirm the root inode matches expectations
	root, err := GetInode(&fs, InoRoot)
	if err != nil {
		t.Fatalf("GetInode(): unexpected err: %v", err)
	}
	if root.Mode.Type != FileTypeDir {
		t.Fatalf(
			"root inode type: wanted `%d`; found `%d`",
			FileTypeDir,
			root.Mode.Type,
		)
	}
	if root.Size != fs.Superblock.BlockSize {
		t.Fatalf(
			"root inode size: wanted `%d`; found `%d`",
			fs.Superblock.BlockSize,
			root.Size,
		)
	}
	if root.LinksCount != 2 {
		t.Fatalf(
			"root inode links count: wanted `2`; found `%d`",
			root.LinksCount,
		)
	}
}
