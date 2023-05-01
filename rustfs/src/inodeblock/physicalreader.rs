use super::{IndirectionError, Result};
use crate::annotate;
use crate::block::{Block, BlockListIndex, Indirection};
use crate::indirectmanager::IndirectReader;
use crate::inode::Inode;

pub struct PhysicalReader<R> {
    indirect_reader: R,
}

impl<R: IndirectReader> PhysicalReader<R> {
    pub fn new(indirect_reader: R) -> PhysicalReader<R> {
        Self { indirect_reader }
    }

    pub fn read_physical(&self, inode: &Inode, inode_block: Block) -> Result<Option<Block>> {
        annotate!(
            match Indirection::from(inode_block) {
                Indirection::Direct { direct_index } => {
                    Ok(direct_index.get_direct_block(&inode.direct_blocks))
                }
                Indirection::SinglyIndirect {
                    singly_indirect_index,
                } => self.get_singly_indirect(inode.singly_indirect_block, singly_indirect_index),
                Indirection::DoublyIndirect {
                    doubly_indirect_index,
                    singly_indirect_index,
                } => self.get_doubly_indirect(
                    inode.doubly_indirect_block,
                    doubly_indirect_index,
                    singly_indirect_index,
                ),
                Indirection::TriplyIndirect {
                    triply_indirect_index,
                    doubly_indirect_index,
                    singly_indirect_index,
                } => self.get_triply_indirect(
                    inode.triply_indirect_block,
                    triply_indirect_index,
                    doubly_indirect_index,
                    singly_indirect_index,
                ),
                Indirection::OutOfRange => Err(IndirectionError::OutOfRange),
            },
            "reading physical block from inode `{}` block `{}`",
            inode.ino,
            inode_block,
        )
    }

    fn get_singly_indirect(
        &self,
        block: Option<Block>,
        index: BlockListIndex,
    ) -> Result<Option<Block>> {
        block.map_or(Ok(None), |b| {
            self.indirect_reader
                .read_indirect(b, index)
                .map_err(IndirectionError::from)
        })
    }

    fn get_doubly_indirect(
        &self,
        block: Option<Block>,
        doubly_indirect_index: BlockListIndex,
        singly_indirect_index: BlockListIndex,
    ) -> Result<Option<Block>> {
        let block = block.map_or(Ok(None), |b| {
            self.indirect_reader.read_indirect(b, doubly_indirect_index)
        })?;
        return self.get_singly_indirect(block, singly_indirect_index);
    }

    fn get_triply_indirect(
        &self,
        block: Option<Block>,
        triply_indirect_index: BlockListIndex,
        doubly_indirect_index: BlockListIndex,
        singly_indirect_index: BlockListIndex,
    ) -> Result<Option<Block>> {
        let block = block.map_or(Ok(None), |b| {
            self.indirect_reader.read_indirect(b, triply_indirect_index)
        })?;
        self.get_doubly_indirect(block, doubly_indirect_index, singly_indirect_index)
    }
}
