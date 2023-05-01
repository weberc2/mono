package store

// allocator implements a simple allocation pool that can grow up to a fixed
// capacity. It is assumed that it will never shrink.
type allocator struct {
	length int
	pool   []entry
}

func newAllocator(capacity int) allocator {
	return allocator{length: 0, pool: make([]entry, capacity)}
}

func (a *allocator) alloc() *entry {
	if a.length >= len(a.pool) {
		return nil
	}
	ret := &a.pool[a.length]
	a.length++
	return ret
}
