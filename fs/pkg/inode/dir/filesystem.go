package dir

import (
	"github.com/weberc2/mono/fs/pkg/alloc"
	"github.com/weberc2/mono/fs/pkg/inode/data"
	"github.com/weberc2/mono/fs/pkg/inode/data/block"
	"github.com/weberc2/mono/fs/pkg/inode/data/block/indirect"
	"github.com/weberc2/mono/fs/pkg/inode/data/block/physical"
	"github.com/weberc2/mono/fs/pkg/io"
	. "github.com/weberc2/mono/fs/pkg/types"
)

type FileSystem struct {
	ReadWriter     data.ReadWriter
	InodeStore     InodeStore
	InoAllocator   alloc.InoAllocator
	BlockAllocator alloc.BlockAllocator
}

func (fs *FileSystem) Init(
	blockAllocator alloc.BlockAllocator,
	inoAllocator alloc.InoAllocator,
	blockVolume io.Volume,
	inodeStore InodeStore,
) {
	indirectReadWriter := indirect.NewReadWriter(blockVolume)
	physicalReadWriter := physical.NewReadWriter(
		blockAllocator,
		indirectReadWriter,
		inodeStore,
	)
	blockReadWriter := block.NewReadWriter(
		physicalReadWriter,
		blockVolume,
	)
	readWriter := data.NewReadWriter(blockReadWriter, inodeStore)
	*fs = FileSystem{
		ReadWriter:     readWriter,
		InodeStore:     inodeStore,
		InoAllocator:   inoAllocator,
		BlockAllocator: blockAllocator,
	}
}
