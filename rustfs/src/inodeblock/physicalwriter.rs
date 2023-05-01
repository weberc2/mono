use super::{IndirectionError, Result};
use crate::bitmap::Bitmap;
use crate::block::*;
use crate::byte::Byte;
use crate::indirectmanager::*;
use crate::inode::*;
use crate::inodestore::InodeStore;

pub struct PhysicalWriter<M, S> {
    pub indirect_manager: M,
    pub inode_store: S,
    pub block_allocator: Bitmap<Block>,
}

impl<M: IndirectReader + IndirectWriter, S: InodeStore> PhysicalWriter<M, S> {
    pub fn new(
        indirect_manager: M,
        inode_store: S,
        allocator: Bitmap<Block>,
    ) -> PhysicalWriter<M, S> {
        PhysicalWriter {
            indirect_manager: indirect_manager,
            inode_store: inode_store,
            block_allocator: allocator,
        }
    }
}

impl<M: IndirectReader + IndirectWriter, S: InodeStore> PhysicalWriter<M, S> {
    pub fn write_physical(
        &mut self,
        inode: &mut Inode,
        inode_block: Block,
        physical_block: Option<Block>,
    ) -> Result<()> {
        match Indirection::from(inode_block) {
            Indirection::Direct { direct_index } => self.update_inode(inode, |inode| {
                inode.set_direct_block(direct_index, physical_block)
            }),
            Indirection::SinglyIndirect {
                singly_indirect_index,
            } => {
                match inode.singly_indirect_block {
                    // if the inode doesn't have a singly indirect block,
                    // allocate one, write the physical block's address into
                    // it, write the newly allocated singly indirect block
                    // address into the inode, and write it back to the inode
                    // store.
                    None => {
                        let singly_indirect_block = alloc(self)?;
                        let allocated_size = BLOCK_SIZE;
                        self.indirect_manager.write_indirect(
                            singly_indirect_block,
                            singly_indirect_index,
                            physical_block,
                        )?;
                        self.update_inode(inode, |inode| {
                            inode.singly_indirect_block = Some(singly_indirect_block);
                            inode.size += allocated_size;
                        })?;
                    }

                    // if the inode's singly indirect block pointer is valid
                    // then we only need to write the physical block address
                    // into it.
                    Some(singly_indirect_block) => self.indirect_manager.write_indirect(
                        singly_indirect_block,
                        singly_indirect_index,
                        physical_block,
                    )?,
                };
                Ok(())
            }
            Indirection::DoublyIndirect {
                doubly_indirect_index,
                singly_indirect_index,
            } => {
                match inode.doubly_indirect_block {
                    None => {
                        // since the doubly indirect block doesn't exist, we
                        // need to allocate it. Since we have to allocate the
                        // doubly indirect block, we will also have to allocate
                        // the singly indirect block.
                        let doubly_indirect_block = alloc(self)?;
                        let singly_indirect_block = alloc(self)?;
                        let allocated_size = BLOCK_SIZE * Byte::new(2);

                        // write the physical block address into the singly
                        // indirect block.
                        self.indirect_manager.write_indirect(
                            singly_indirect_block,
                            singly_indirect_index,
                            physical_block,
                        )?;

                        // write the singly indirect block address into the
                        // doubly indirect block.
                        self.indirect_manager.write_indirect(
                            doubly_indirect_block,
                            doubly_indirect_index,
                            Some(singly_indirect_block),
                        )?;

                        // update the clone's doubly indirect block pointer
                        // field.
                        let mut clone = inode.clone();
                        clone.doubly_indirect_block = Some(doubly_indirect_block);
                        clone.size += allocated_size;
                        self.inode_store.put(&clone)?;
                        *inode = clone;
                    }
                    Some(doubly_indirect_block) => {
                        // since the doubly indirect block already exists, we
                        // should check to see if the singly indirect block
                        // also already exists.
                        match self
                            .indirect_manager
                            .read_indirect(doubly_indirect_block, doubly_indirect_index)?
                        {
                            // if the singly indirect block doesn't already
                            // exist, allocate it, write the physical block
                            // address into the singly indirect block, and
                            // write the singly indirect block's address into
                            // the doubly indirect block.
                            None => {
                                let singly_indirect_block = alloc(self)?;
                                let allocated_size = BLOCK_SIZE;
                                self.indirect_manager.write_indirect(
                                    singly_indirect_block,
                                    singly_indirect_index,
                                    physical_block,
                                )?;
                                self.indirect_manager.write_indirect(
                                    doubly_indirect_block,
                                    doubly_indirect_index,
                                    Some(singly_indirect_block),
                                )?;
                                self.update_inode(inode, |inode| inode.size += allocated_size)?;
                            }

                            // since the singly indirect block already exists,
                            // all we need to do is write the physical block's
                            // address into the singly indirect block.
                            Some(singly_indirect_block) => self.indirect_manager.write_indirect(
                                singly_indirect_block,
                                singly_indirect_index,
                                physical_block,
                            )?,
                        }
                    }
                }

                Ok(())
            }
            Indirection::TriplyIndirect {
                triply_indirect_index,
                doubly_indirect_index,
                singly_indirect_index,
            } => match inode.triply_indirect_block {
                None => {
                    let triply_indirect_block = alloc(self)?;
                    let doubly_indirect_block = alloc(self)?;
                    let singly_indirect_block = alloc(self)?;
                    let allocated_size = BLOCK_SIZE * Byte::new(3);

                    self.indirect_manager.write_indirect(
                        singly_indirect_block,
                        singly_indirect_index,
                        physical_block,
                    )?;
                    self.indirect_manager.write_indirect(
                        doubly_indirect_block,
                        doubly_indirect_index,
                        Some(singly_indirect_block),
                    )?;
                    self.indirect_manager.write_indirect(
                        triply_indirect_block,
                        triply_indirect_index,
                        Some(doubly_indirect_block),
                    )?;

                    self.update_inode(inode, |inode| {
                        inode.size += allocated_size;
                        inode.triply_indirect_block = Some(triply_indirect_block);
                    })
                }
                Some(triply_indirect_block) => match self
                    .indirect_manager
                    .read_indirect(triply_indirect_block, triply_indirect_index)?
                {
                    None => {
                        let doubly_indirect_block = alloc(self)?;
                        let singly_indirect_block = alloc(self)?;
                        let allocated_size = BLOCK_SIZE * Byte::new(2);

                        self.indirect_manager.write_indirect(
                            singly_indirect_block,
                            singly_indirect_index,
                            physical_block,
                        )?;
                        self.indirect_manager.write_indirect(
                            doubly_indirect_block,
                            doubly_indirect_index,
                            Some(singly_indirect_block),
                        )?;
                        self.indirect_manager.write_indirect(
                            triply_indirect_block,
                            triply_indirect_index,
                            Some(doubly_indirect_block),
                        )?;

                        self.update_inode(inode, |inode| inode.size += allocated_size)
                    }
                    Some(doubly_indirect_block) => match self
                        .indirect_manager
                        .read_indirect(doubly_indirect_block, doubly_indirect_index)?
                    {
                        // Since the triply and doubly (but not singly!)
                        // indirect blocks have been allocated/prepared, we
                        // only have to allocate the singly indirect block (and
                        // thus update the inode's size), write the physical
                        // block address to the singly indirect block, and
                        // write the singly indirect block's address to the
                        // doubly indirect block.
                        None => {
                            let singly_indirect_block = alloc(self)?;
                            let allocated_size = BLOCK_SIZE;

                            self.indirect_manager.write_indirect(
                                singly_indirect_block,
                                singly_indirect_index,
                                physical_block,
                            )?;
                            self.indirect_manager.write_indirect(
                                doubly_indirect_block,
                                doubly_indirect_index,
                                Some(singly_indirect_block),
                            )?;

                            self.update_inode(inode, |inode| inode.size += allocated_size)
                        }

                        // Since the triply, doubly, and singly indirect blocks
                        // are all allocated, there is no need to update the
                        // inode. All we have to do is write the physical block
                        // address to the singly indirect block.
                        Some(singly_indirect_block) => self
                            .indirect_manager
                            .write_indirect(
                                singly_indirect_block,
                                singly_indirect_index,
                                physical_block,
                            )
                            .map_err(IndirectionError::from),
                    },
                },
            },
            Indirection::OutOfRange => Err(IndirectionError::OutOfRange),
        }
    }

    // Transactionally update the inode. This isn't atomic in the sense that
    // it's threadsafe, but rather that errors updating the inode store won't
    // leave the inode object in a different state than the same inode in the
    // inode store.
    fn update_inode<F>(&mut self, inode: &mut Inode, mut f: F) -> Result<()>
    where
        F: FnMut(&mut Inode),
    {
        let mut clone = inode.clone();
        f(&mut clone);
        self.inode_store.put(&clone)?;
        *inode = clone;
        Ok(())
    }
}

// public so we can access this via tests; do not export outside of the parent
// module.
pub fn alloc<M, S>(c: &mut PhysicalWriter<M, S>) -> Result<Block> {
    c.block_allocator
        .allocate()
        .ok_or(IndirectionError::OutOfBlocks)
}

pub trait Allocator {
    fn allocate(&self) -> Option<Block>;
}
