use crate::byte::Byte;
use crate::inode::{Ino, Inode};
use crate::inodestore::InodeStore;
use crate::io::{ReadAt, WriteAt};
use crate::write::{read_inode, write_inode};
use std::io::Result;

pub struct VolumeInodeStore<V> {
    volume: V,
}

impl<V> VolumeInodeStore<V> {
    pub fn new(volume: V) -> VolumeInodeStore<V> {
        VolumeInodeStore { volume }
    }
}

impl<V: ReadAt + WriteAt> InodeStore for VolumeInodeStore<V> {
    fn get(&mut self, ino: Ino) -> Result<Option<Inode>> {
        Ok(Some(read_inode(&mut self.volume, ino)?))
    }

    fn put(&mut self, inode: &Inode) -> Result<()> {
        write_inode(&mut self.volume, inode)
    }
}
