package alloc

import . "github.com/weberc2/mono/fs/pkg/types"

type InoAllocator struct {
	Allocator
}

func (ia InoAllocator) Alloc() (Ino, bool) {
	if ino, ok := ia.Allocator.Alloc(); ok {
		return Ino(ino + 1), true
	}
	return InoNil, false
}

func (ia InoAllocator) Free(ino Ino) {
	ia.Allocator.Free(uint64(ino) - 1)
}

func (ia InoAllocator) Reserve(ino Ino) {
	ia.Allocator.Reserve(uint64(ino) - 1)
}
