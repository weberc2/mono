package fs

type Cache struct {
	head      *entry
	tail      *entry
	byIno     map[Ino]*entry
	allocator allocator
}

func NewCache(capacity int) Cache {
	return Cache{
		head:      nil,
		tail:      nil,
		byIno:     make(map[Ino]*entry, capacity),
		allocator: newAllocator(capacity),
	}
}

func (c *Cache) moveToFront(e *entry) {
	// only do this if this isn't the first node in the cache (otherwise
	// `c.head` will be nil)
	if c.head != nil {
		c.head.prev = e
	}
	e.next = c.head
	c.head = e
}

func (c *Cache) Push(value Inode) (evicted *Inode) {
	current, exists := c.byIno[value.Ino]
	if !exists {
		current = c.allocator.alloc()

		// if the allocator is at capacity, pop the tail and store it in `current`
		if current == nil {
			// copy the value onto the heap to return later; we don't return
			// pointers into the cache
			evicted = new(Inode)
			*evicted = c.tail.value

			current = c.tail
			c.tail = c.tail.prev
			c.tail.next = nil
			current.prev = nil
			delete(c.byIno, current.value.Ino)
		}

		// since it's not in the `byIno` map, add it
		c.byIno[value.Ino] = current
	}

	c.moveToFront(current)
	c.head.value = value
	return
}

func (c *Cache) Get(ino Ino) (Inode, bool) {
	e := c.byIno[ino]
	if e == nil {
		return Inode{}, false
	}

	// unlink the entry associated with `ino` and make it the new head since it
	// was most recently used.
	if e.prev != nil {
		e.prev.next = e.next
	}
	if e.next != nil {
		e.next.prev = e.prev
	}
	e.next = c.head
	e.prev = nil
	c.head = e
	return e.value, true
}

func (c *Cache) Remove(ino Ino) (Inode, bool) {
	e := c.byIno[ino]
	if e == nil {
		return Inode{}, false
	}

	// unlink the entry associated with `ino`, move it to least-recently-used,
	// and associate it with an invalid ino
	if e != c.head {
		e.prev.next = e.next
	}
	if e != c.tail {
		e.next.prev = e.prev
	}
	e.next = nil
	e.prev = c.tail
	c.tail = e
	ret := e.value
	e.value = Inode{Ino: InoOutOfInodes}
	return ret, true
}

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

type entry struct {
	next  *entry
	prev  *entry
	value Inode
}
