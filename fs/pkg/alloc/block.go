package alloc

import . "github.com/weberc2/mono/fs/pkg/types"

type BlockAllocator struct {
	Allocator
}

func (ba BlockAllocator) Alloc() (Block, bool) {
	if b, ok := ba.Allocator.Alloc(); ok {
		return Block(b + 1), true
	}
	return BlockNil, false
}

func (ba BlockAllocator) Free(b Block) {
	ba.Allocator.Free(uint64(b) - 1)
}

func (ba BlockAllocator) Reserve(b Block) {
	ba.Allocator.Reserve(uint64(b) - 1)
}
