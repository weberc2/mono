package store

import (
	. "github.com/weberc2/mono/fs/pkg/types"
)

type Cache struct {
	head      *entry
	tail      *entry
	lookup    map[Ino]*entry
	allocator allocator
}

func NewCache(capacity int) *Cache {
	return &Cache{
		lookup:    make(map[Ino]*entry),
		allocator: newAllocator(capacity),
	}
}

func (c *Cache) Get(ino Ino, out *Inode) bool {
	e, exists := c.lookup[ino]
	if !exists {
		return false
	}

	c.moveFront(e)
	*out = e.value
	return true
}

func (c *Cache) Remove(ino Ino, removed *Inode) bool {
	e := c.lookup[ino]
	if e == nil {
		return false
	}

	c.unlink(e)

	// fetch the return value from the entry and zero out the entry
	*removed = e.value
	e.value = Inode{}

	// move the wiped entry to the tail
	e.next = nil
	e.prev = c.tail
	c.tail.next = e
	c.tail = e
	return true
}

func (c *Cache) unlink(e *entry) {
	// if e is not the head, then it will have a non-nil `prev` field; we need
	// to set that `prev` to point to `e.next`
	if e != c.head {
		e.prev.next = e.next
	}

	// similarly, if e is not the tail, then it will have a non-nil `next`
	// field; we need to set `e.next`'s `prev` to point to `e.prev`
	if e != c.tail {
		e.next.prev = e.prev
	}
}

func (c *Cache) Push(inode *Inode, evicted *Inode) (evict bool) {
	if e, exists := c.lookup[inode.Ino]; exists {
		c.moveFront(e)
		e.value = *inode
		return false
	}

	e := c.allocator.alloc()
	if e == nil {
		// If the allocator's capacity is 0, then we could have a case where
		// c.tail is nil and the allocator returns nil, but this is a niche
		// programming error and it's okay to let it blow up in a nil pointer
		// exception.
		*evicted = c.tail.value
		delete(c.lookup, evicted.Ino)
		e = c.tail
		evict = true
	} else if c.tail == nil {
		// if the tail is nil, set it to the new entry

		// NB: c.tail cannot be nil if allocation failed unless the allocator's
		// capacity is zero (this is invalid / not supported).
		c.tail = e
	}

	e.value = *inode
	c.lookup[inode.Ino] = e
	c.moveFront(e)
	return
}

func (c *Cache) moveFront(e *entry) {
	if c.head != nil {
		c.head.prev = e
	}
	e.next = c.head
	e.prev = nil
	c.head = e
}

type entry struct {
	prev  *entry
	next  *entry
	value Inode
}
