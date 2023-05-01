package store

import (
	"fmt"

	. "github.com/weberc2/mono/fs/pkg/types"
)

type CachingInodeStore struct {
	backend InodeStore
	cache   Cache
	dirty   map[Ino]struct{}
}

func NewCachingInodeStore(
	backend InodeStore,
	cacheCapacity int,
) *CachingInodeStore {
	return &CachingInodeStore{
		backend: backend,
		cache:   *NewCache(cacheCapacity),
		dirty:   make(map[Ino]struct{}),
	}
}

func (store *CachingInodeStore) Put(inode *Inode) error {
	var evicted Inode
	store.dirty[inode.Ino] = struct{}{}
	if store.cache.Push(inode, &evicted) {
		if err := store.backend.Put(&evicted); err != nil {
			return fmt.Errorf(
				"storing inode `%d`: flushing evicted inode `%d` to disc: %w",
				inode.Ino,
				evicted.Ino,
				err,
			)
		}
		delete(store.dirty, evicted.Ino)
	}
	return nil
}

func (store *CachingInodeStore) Get(ino Ino, output *Inode) error {
	if store.cache.Get(ino, output) {
		return nil
	}

	if err := store.backend.Get(ino, output); err != nil {
		return fmt.Errorf(
			"fetching inode `%d`: cache miss; checking backend store: %w",
			ino,
			err,
		)
	}

	return nil
}

func (store *CachingInodeStore) Flush(ino Ino) error {
	var removed Inode
	// Remove the ino from the cache; if it was removed *and* it was dirty,
	// then write it to the backend volume and remove it from the dirty set.
	if store.cache.Remove(ino, &removed) {
		if _, dirty := store.dirty[ino]; dirty {
			if err := store.backend.Put(&removed); err != nil {
				return fmt.Errorf("flushing inode `%d`: %w", removed.Ino, err)
			}
			delete(store.dirty, ino)
		}
	}
	return nil
}
