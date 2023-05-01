package fs

import (
	"encoding/json"
	"io"
	"testing"
)

func TestSuperblockEncodeDecode(t *testing.T) {
	wanted := NewSuperblock(1024, 1024, 50)
	buf := [SuperblockSize]byte{}
	EncodeSuperblock(&wanted, &buf)
	var found Superblock
	if err := DecodeSuperblock(&found, &buf); err != nil {
		t.Fatalf("DecodeSuperblock(): unexpected err: %v", err)
	}
	wantedData, err := json.Marshal(&wanted)
	if err != nil {
		t.Fatalf("marshaling `wanted` Superblock: %v", err)
	}
	foundData, err := json.Marshal(&found)
	if err != nil {
		t.Fatalf("marshaling `found` Superblock: %v", err)
	}
	if wanted != found {
		t.Fatalf(
			"DecodeSuperblock(): wanted `%s`; found `%s`",
			wantedData,
			foundData,
		)
	}
}

func TestSuperblockReadWrite(t *testing.T) {
	buf := NewBuffer(nil)
	fsWrite := NewFileSystem(&FileSystemParams{
		Volume:        buf,
		BlockSize:     DefaultBlockSize,
		Blocks:        1024,
		Inodes:        50,
		CacheCapacity: 10,
	})

	if err := WriteSuperblock(&fsWrite); err != nil {
		t.Fatalf("writing superblock: %v", err)
	}

	_, _ = buf.Seek(0, io.SeekStart)
	fsRead := FileSystem{Volume: buf}
	if err := ReadSuperblock(&fsRead); err != nil {
		t.Fatalf("reading superblock: %v", err)
	}

	if fsRead.Superblock != fsWrite.Superblock {
		wrote, err := json.Marshal(&fsWrite.Superblock)
		if err != nil {
			t.Fatalf("marshaling 'wrote' superblock: %v", err)
		}
		read, err := json.Marshal(&fsRead.Superblock)
		if err != nil {
			t.Fatalf("marshaling 'read' superblock: %v", err)
		}
		t.Fatalf("wrote superblock `%s`; read superblock `%s`", wrote, read)
	}
}
