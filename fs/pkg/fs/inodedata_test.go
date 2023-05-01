package fs

import (
	"bytes"
	"testing"
)

func TestInodeData(t *testing.T) {
	fs := NewFileSystem(&FileSystemParams{
		Volume:        NewBuffer(make([]byte, 1024*1024)),
		BlockSize:     DefaultBlockSize,
		Blocks:        100,
		Inodes:        50,
		CacheCapacity: 10,
	})
	if err := fs.Init(); err != nil {
		t.Fatalf("Init(): unexpected err: %v", err)
	}

	// NB: even though InoRoot points to a directory, for low level operations
	// like reading and writing inode data, it doesn't matter; however, it's
	// easier than creating a FileTypeRegular inode (making an inode in root
	// depends on reading and writing inode data--which we're testing here--so
	// we end up in a dependency cycle).
	inode, err := GetInode(&fs, InoRoot)
	if err != nil {
		t.Fatalf("GetInode(root): unexpected err: %v", err)
	}

	wanted := []byte("hello")
	if _, err := WriteInodeData(&fs, &inode, 0, wanted); err != nil {
		t.Fatalf("WriteInodeData(): unexpected err: %v", err)
	}

	found := make([]byte, 5)
	if _, err := ReadInodeData(&fs, &inode, 0, found); err != nil {
		t.Fatalf("ReadInodeData(): unexpected err: %v", err)
	}

	if !bytes.Equal(wanted, found) {
		t.Fatalf("ReadInodeData(): wanted `%s`; found `%s`", wanted, found)
	}
}
