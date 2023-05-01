use crate::block::{Block, BLOCK_SIZE};
use crate::byte::Byte;
use crate::indirectmanager::IndirectReader;
use crate::inode::Inode;
use crate::inodeblock::{IndirectionError, Manager, Reader, Result, Writer};
use crate::inodestore::InodeStore;
use crate::io::{ReadAt, WriteAt};
use std::cmp;

pub struct InodeDataManager<S: InodeStore, M: Reader + Writer> {
    block_manager: M,
    inode_store: S,
}

impl<S: InodeStore, M: Reader + Writer> InodeDataManager<S, M> {
    pub fn new(block_manager: M, inode_store: S) -> InodeDataManager<S, M> {
        InodeDataManager {
            block_manager,
            inode_store,
        }
    }

    pub fn read_inode_data(
        &mut self,
        inode: &Inode,
        offset: Byte,
        buffer: &mut [u8],
    ) -> Result<Byte> {
        let max_length = Byte::from(cmp::min(
            buffer.len() as u64,
            u64::from(inode.size - offset),
        ));
        let mut chunk_begin = Byte::new(0);
        while chunk_begin < max_length {
            let chunk_block = Block::from(u64::from((offset + chunk_begin) / BLOCK_SIZE));
            let chunk_offset = (offset + chunk_begin) % BLOCK_SIZE;
            let chunk_length = Byte::from(cmp::min(
                max_length - chunk_begin,
                BLOCK_SIZE - chunk_offset,
            ));

            self.block_manager.read_inode_block(
                inode,
                chunk_block,
                chunk_offset,
                &mut buffer[usize::from(chunk_begin)..usize::from(chunk_length)],
            )?;
            chunk_begin = chunk_begin + chunk_length;
        }
        Ok(chunk_begin)
    }

    pub fn write_inode_data(
        &mut self,
        inode: &mut Inode,
        offset: Byte,
        buffer: &[u8],
    ) -> Result<Byte> {
        let mut chunk_begin = Byte::new(0);
        while chunk_begin < Byte::from(buffer.len()) {
            let chunk_block = Block::from(u64::from((offset + chunk_begin) / BLOCK_SIZE));
            let chunk_offset = (offset + chunk_begin) % BLOCK_SIZE;
            let chunk_length = cmp::min(
                Byte::from(buffer.len()) - chunk_begin,
                BLOCK_SIZE - chunk_offset,
            );
            self.block_manager.write_inode_block(
                inode,
                chunk_block,
                chunk_offset,
                &buffer[usize::from(chunk_begin)..usize::from(chunk_length)],
            )?;
            chunk_begin = chunk_begin + chunk_length;
        }

        if inode.size < offset + chunk_begin {
            inode.size = offset + chunk_begin;
            crate::annotate!(
                self.inode_store.put(inode),
                "putting inode{0} into inode store",
                inode.ino
            )?;
        }

        Ok(chunk_begin)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::inodestore::InodeStore;
    use crate::{indirectmanager, inode::Ino, inodeblock::PhysicalReader};
    use std::collections::HashMap;

    #[test]
    fn test() -> Result<()> {
        let mut inode = Inode::new(Ino::new(0));
        inode.direct_blocks[0] = Some(Block::new(0));

        let volume = vec![0 as u8; 1024 * 1024];
        let physical_reader = PhysicalReader::new(indirectmanager::HashMap::new());
        let mut block_manager = Manager::new(physical_reader, volume);
        let mut inode_store = HashMap::from([(inode.ino, inode.clone())]);

        {
            let buffer = [1, 1, 1, 1];
            let mut data_manager = InodeDataManager::new(&mut block_manager, &mut inode_store);
            data_manager.write_inode_data(&mut inode, Byte::new(0), &buffer)?;

            // make sure the inode's size was updated
            assert_eq!(Byte::from(buffer.len()), inode.size);

            // make sure the inode store's entry was updated
            assert_eq!(
                <HashMap<Ino, Inode> as InodeStore>::get(&mut inode_store, inode.ino)?,
                Some(inode.clone())
            );
        }

        {
            let mut buffer = [0; 4];
            let mut data_manager = InodeDataManager::new(&mut block_manager, &mut inode_store);
            data_manager.read_inode_data(&inode, Byte::new(0), &mut buffer)?;
            assert_eq!([1, 1, 1, 1], buffer);
        }
        Ok(())
    }
}
