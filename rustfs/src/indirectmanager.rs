use crate::block::*;
use crate::io::{ReadAt, WriteAt};
use std::cell::Cell;
use std::io::Result;

pub trait IndirectReader {
    fn read_indirect(&self, indirect_block: Block, index: BlockListIndex) -> Result<Option<Block>>;
}

pub trait IndirectWriter {
    fn write_indirect(
        &mut self,
        indirect_block: Block,
        index: BlockListIndex,
        block: Option<Block>,
    ) -> Result<()>;
}

pub trait IndirectManager: IndirectReader + IndirectWriter {}

pub type HashMap = std::collections::HashMap<(Block, BlockListIndex), Option<Block>>;

impl IndirectReader for HashMap {
    fn read_indirect(&self, indirect: Block, index: BlockListIndex) -> Result<Option<Block>> {
        match self.get(&(indirect, index)) {
            Some(target) => Ok(*target),
            None => Ok(None),
        }
    }
}

impl IndirectWriter for HashMap {
    fn write_indirect(
        &mut self,
        indirect: Block,
        index: BlockListIndex,
        target: Option<Block>,
    ) -> std::io::Result<()> {
        self.insert((indirect, index), target);
        Ok(())
    }
}

#[derive(Clone, Copy)]
pub struct Wrapper<T> {
    inner: T,
}

impl<T> Wrapper<T> {
    pub fn new(inner: T) -> Self {
        Self { inner }
    }
}

impl<V: crate::io::ReadAt> IndirectReader for Wrapper<V> {
    fn read_indirect(&self, indirect: Block, index: BlockListIndex) -> Result<Option<Block>> {
        use byteorder::{ByteOrder, LittleEndian};
        let mut buf = [0; BLOCK_POINTER_SIZE.to_usize()];
        self.inner.read_at(offset(indirect, index), &mut buf)?;
        Ok(Block::decode(LittleEndian::read_u64(&buf)))
    }
}

impl<V: crate::io::WriteAt> IndirectWriter for Wrapper<V> {
    fn write_indirect(
        &mut self,
        indirect_block: Block,
        index: BlockListIndex,
        block: Option<Block>,
    ) -> Result<()> {
        use byteorder::{ByteOrder, LittleEndian};
        let mut buf = [0; BLOCK_POINTER_SIZE.to_usize()];
        LittleEndian::write_u64(&mut buf, Block::encode(block));
        self.inner.write_at(offset(indirect_block, index), &mut buf)
    }
}

fn offset(indirect_block: Block, index: BlockListIndex) -> crate::byte::Byte {
    let offset_of_block = crate::byte::Byte::new(u64::from(indirect_block)) * BLOCK_SIZE;
    let offset_in_block = crate::byte::Byte::new(u64::from(index)) * BLOCK_POINTER_SIZE;
    offset_of_block + offset_in_block
}
