use crate::inode::*;
use crate::inodestore::InodeStore;
use lru::LruCache;
use std::num::NonZeroUsize;

pub struct CachingInodeStore<S: InodeStore> {
    backend: S,
    cache: LruCache<Ino, Inode>,
}

impl<S: InodeStore> CachingInodeStore<S> {
    pub fn new(backend: S, capacity: NonZeroUsize) -> CachingInodeStore<S> {
        CachingInodeStore {
            backend: backend,
            cache: LruCache::new(capacity),
        }
    }
}

impl<S: InodeStore> InodeStore for CachingInodeStore<S> {
    fn put(&mut self, inode: &Inode) -> std::io::Result<()> {
        match self.cache.push(inode.ino, inode.clone()) {
            None => Ok(()),

            // `lru::LruCache::push` will return a non-`None` value if the
            // cache was full *or* if the cache already contained the key. In
            // the latter case, we don't want to write to disk, particularly
            // since the resulting inode would be out of date.
            Some((ino, popped)) => match ino == inode.ino {
                // do nothing; `popped` is garbage
                true => Ok(()),

                // `popped` is the least-recently used value; write it to disk.
                false => self.backend.put(&popped),
            },
        }
    }

    fn get(&mut self, ino: Ino) -> std::io::Result<Option<Inode>> {
        match self.cache.get(&ino) {
            Some(inode) => Ok(Some(inode.clone())),
            None => match self.backend.get(ino)? {
                Some(inode) => {
                    self.put(&inode)?;
                    Ok(Some(inode))
                }
                None => Ok(None),
            },
        }
    }
}
