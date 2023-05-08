package alloc

type Allocator interface {
	Alloc() (uint64, bool)
	Reserve(uint64)
	Free(uint64)
}

var (
	_ Allocator = (*Bitmap)(nil)
	_ Allocator = (*FlushableBitmap)(nil)
)
