package data

import (
	"testing"

	"github.com/weberc2/mono/fs/pkg/alloc"
	"github.com/weberc2/mono/fs/pkg/inode/data/block"
	"github.com/weberc2/mono/fs/pkg/inode/data/block/indirect"
	"github.com/weberc2/mono/fs/pkg/inode/data/block/physical"
	"github.com/weberc2/mono/fs/pkg/inode/store"
	"github.com/weberc2/mono/fs/pkg/io"
	. "github.com/weberc2/mono/fs/pkg/types"
)

func TestWrite(t *testing.T) {
	volume := io.NewBuffer(make([]byte, 1024*10))
	indirectReadWriter := indirect.NewReadWriter(volume)
	bitmap := alloc.New(1024)
	offsetVolume := io.NewOffsetVolume(volume, BlockSize)
	backendInodeStore := store.NewVolumeInodeStore(offsetVolume)
	inodeStore := store.NewCachingInodeStore(backendInodeStore, 10)
	physicalReadWriter := physical.NewReadWriter(
		alloc.BlockAllocator{Allocator: &bitmap},
		indirectReadWriter,
		inodeStore,
	)
	blockWriter := block.NewWriter(physicalReadWriter, volume)
	writer := NewWriter(blockWriter, inodeStore)

	inode := Inode{Ino: 0, FileType: FileTypeDir}
	input := []byte("hello")
	n, err := writer.Write(&inode, 0, input)
	if err != nil {
		t.Fatalf("Write(): unexpected err: %v", err)
	}

	if n != Byte(len(input)) {
		t.Fatalf(
			"Write(): wanted `%d` bytes written; found `%d`",
			len(input),
			n,
		)
	}
}
