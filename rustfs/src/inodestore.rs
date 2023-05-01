use crate::inode::{Ino, Inode};
use std::io::Result;

pub trait InodeStore {
    fn put(&mut self, inode: &Inode) -> Result<()>;
    fn get(&mut self, ino: Ino) -> Result<Option<Inode>>;
}

impl<T: InodeStore> InodeStore for &mut T {
    fn put(&mut self, inode: &Inode) -> std::io::Result<()> {
        (**self).put(inode)
    }

    fn get(&mut self, ino: Ino) -> std::io::Result<Option<Inode>> {
        (**self).get(ino)
    }
}

impl InodeStore for std::collections::HashMap<Ino, Inode> {
    fn get(&mut self, ino: Ino) -> Result<Option<Inode>> {
        Ok(std::collections::HashMap::get(self, &ino).map(Inode::clone))
    }

    fn put(&mut self, inode: &Inode) -> Result<()> {
        let _ = self.insert(inode.ino, inode.clone());
        Ok(())
    }
}
