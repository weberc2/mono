use super::physicalwriter::*;
use super::*;
use crate::bitmap::Bitmap;
use crate::block::{
    Block, BlockListIndex, BLOCK_POINTERS_PER_BLOCK, BLOCK_SIZE, DIRECT_BLOCKS_COUNT,
    DIRECT_BLOCKS_MAX, DOUBLY_INDIRECT_MAX, SINGLY_INDIRECT_MAX, TRIPLY_INDIRECT_MAX,
};
use crate::byte::Byte;
use crate::indirectmanager::{IndirectManager, IndirectReader, IndirectWriter};
use crate::inode::{Ino, Inode};
use std::collections::HashMap;

#[test]
fn test_write_direct() -> Result<()> {
    let mut converter = PhysicalWriter::default();
    let mut inode = Inode::new(Ino::new(0));

    converter.write_physical(&mut inode, Block::new(0), Some(Block::new(1)))?;

    let mut wanted = Inode::new(Ino::new(0));
    wanted.direct_blocks[0] = Some(Block::new(1));
    assert_eq!(wanted, inode);

    let mut wanted_inodes = HashMap::new();
    wanted_inodes.insert(wanted.ino, wanted);
    assert_eq!(wanted_inodes, converter.inode_store);

    Ok(())
}

#[test]
fn test_write_singly_indirect_allocate() -> Result<()> {
    let mut converter = PhysicalWriter::default();
    let mut inode = Inode::new(Ino::new(0));

    // When we try to set an inode block that is beneath the doubly-indirect
    // range but above the direct range
    let inode_block = DIRECT_BLOCKS_MAX + Block::new(42);
    let physical_block = Some(Block::new(100));
    converter.write_physical(&mut inode, inode_block, physical_block)?;

    // Then expect the inode was updated
    let mut wanted_inode = Inode::new(Ino::new(0));
    let singly_indirect_block = Block::new(0);
    wanted_inode.singly_indirect_block = Some(singly_indirect_block);
    wanted_inode.size = BLOCK_SIZE;
    assert_eq!(wanted_inode, inode);

    // And expect the indirect converter was updated
    let mut wanted_table = HashMap::new();
    let singly_indirect_index = BlockListIndex::new(42);
    wanted_table.insert(
        (singly_indirect_block, singly_indirect_index),
        physical_block,
    );
    assert_eq!(wanted_table, converter.indirect_manager.table);

    Ok(())
}

#[test]
fn test_write_singly_indirect_no_allocate() -> Result<()> {
    let mut converter = PhysicalWriter::default();
    let mut inode = Inode::new(Ino::new(0));

    // Given the singly indirect block is prepared for the inode
    let singly_indirect_block = alloc(&mut converter)?;
    inode.singly_indirect_block = Some(singly_indirect_block);
    inode.size = BLOCK_SIZE;

    // When we try to set an inode block that is beneath the doubly-indirect
    // range but above the direct range
    let inode_block = DIRECT_BLOCKS_MAX + Block::new(42);
    let physical_block = Some(Block::new(100));
    converter.write_physical(&mut inode, inode_block, physical_block)?;

    // Then expect the inode was updated
    let mut wanted_inode = Inode::new(Ino::new(0));
    wanted_inode.singly_indirect_block = Some(singly_indirect_block);
    wanted_inode.size = BLOCK_SIZE;
    assert_eq!(wanted_inode, inode);

    // And expect the indirect converter was updated
    let mut wanted_table = HashMap::new();
    let singly_indirect_index = BlockListIndex::new(42);
    wanted_table.insert(
        (singly_indirect_block, singly_indirect_index),
        physical_block,
    );
    assert_eq!(wanted_table, converter.indirect_manager.table);

    Ok(())
}

#[test]
fn test_write_doubly_indirect_allocate_both() -> Result<()> {
    let mut converter = PhysicalWriter::default();
    let mut inode = Inode::new(Ino::new(0));

    // When we try to set an inode block that is outside of the singly-indirect
    // range but below the minimum triply-indirect
    let inode_block = SINGLY_INDIRECT_MAX + Block::new(42);
    let physical_block = Some(Block::new(100));
    converter.write_physical(&mut inode, inode_block, physical_block)?;

    // Then expect the inode was updated
    let mut wanted_inode = Inode::new(Ino::new(0));
    let doubly_indirect_block = Block::new(0);
    wanted_inode.doubly_indirect_block = Some(doubly_indirect_block);
    wanted_inode.size = BLOCK_SIZE * Byte::new(2);
    assert_eq!(wanted_inode, inode);

    // And expect the indirect converter was updated
    let mut wanted_table = HashMap::new();
    let doubly_indirect_index = BlockListIndex::new(0);
    let singly_indirect_block = Block::new(1);
    let singly_indirect_index = BlockListIndex::new(42);
    wanted_table.insert(
        (doubly_indirect_block, doubly_indirect_index),
        Some(singly_indirect_block),
    );
    wanted_table.insert(
        (singly_indirect_block, singly_indirect_index),
        physical_block,
    );
    assert_eq!(wanted_table, converter.indirect_manager.table);

    Ok(())
}

#[test]
fn test_write_doubly_indirect_allocate_one() -> Result<()> {
    let mut converter = PhysicalWriter::default();
    let mut inode = Inode::new(Ino::new(0));

    // Given the inode's doubly indirect block (and *not* the singly indirect
    // block) has already been prepared
    let doubly_indirect_block = alloc(&mut converter)?;
    inode.doubly_indirect_block = Some(doubly_indirect_block);
    inode.size = BLOCK_SIZE;

    // When we try to set an inode block that is outside of the singly-indirect
    // range but below the minimum triply-indirect
    let inode_block = SINGLY_INDIRECT_MAX + Block::new(42);
    let physical_block = Some(Block::new(100));
    converter.write_physical(&mut inode, inode_block, physical_block)?;

    // Then expect the inode was updated
    let mut wanted_inode = Inode::new(Ino::new(0));
    wanted_inode.doubly_indirect_block = Some(doubly_indirect_block);
    wanted_inode.size = BLOCK_SIZE * Byte::new(2);
    assert_eq!(wanted_inode, inode);

    // And expect the indirect converter was updated
    let mut wanted_table = HashMap::new();
    let doubly_indirect_index = BlockListIndex::new(0);
    let singly_indirect_block = Block::new(1);
    let singly_indirect_index = BlockListIndex::new(42);
    wanted_table.insert(
        (doubly_indirect_block, doubly_indirect_index),
        Some(singly_indirect_block),
    );
    wanted_table.insert(
        (singly_indirect_block, singly_indirect_index),
        physical_block,
    );
    assert_eq!(wanted_table, converter.indirect_manager.table);

    Ok(())
}

#[test]
fn test_write_doubly_indirect_allocate_neither() -> Result<()> {
    let mut converter = PhysicalWriter::default();
    let mut inode = Inode::new(Ino::new(0));

    // Given the inode's doubly and singly indirect blocks have already been
    // prepared
    let doubly_indirect_block = alloc(&mut converter)?;
    let singly_indirect_block = alloc(&mut converter)?;
    let doubly_indirect_index = BlockListIndex::new(0);
    inode.doubly_indirect_block = Some(doubly_indirect_block);
    inode.size = BLOCK_SIZE * Byte::new(2); // 2 blocks have been allocated
    converter.indirect_manager.table.insert(
        (doubly_indirect_block, doubly_indirect_index),
        Some(singly_indirect_block),
    );

    // When we try to set an inode block that is outside of the singly-indirect
    // range but below the minimum triply-indirect
    let inode_block = SINGLY_INDIRECT_MAX + Block::new(42);
    let physical_block = Some(Block::new(100));
    converter.write_physical(&mut inode, inode_block, physical_block)?;

    // Then expect the inode was updated
    let mut wanted_inode = Inode::new(Ino::new(0));
    wanted_inode.doubly_indirect_block = Some(doubly_indirect_block);
    wanted_inode.size = BLOCK_SIZE * Byte::new(2);
    assert_eq!(wanted_inode, inode);

    // And expect the indirect converter was updated
    let mut wanted_table = HashMap::new();
    let doubly_indirect_index = BlockListIndex::new(0);
    let singly_indirect_block = Block::new(1);
    let singly_indirect_index = BlockListIndex::new(42);
    wanted_table.insert(
        (doubly_indirect_block, doubly_indirect_index),
        Some(singly_indirect_block),
    );
    wanted_table.insert(
        (singly_indirect_block, singly_indirect_index),
        physical_block,
    );
    assert_eq!(wanted_table, converter.indirect_manager.table);

    Ok(())
}

#[test]
fn test_write_triply_indirect_allocate_all() -> Result<()> {
    let mut converter = PhysicalWriter::default();
    let mut inode = Inode::new(Ino::new(0));

    // When we try to set an inode block that is outside of the doubly-indirect
    // range
    let inode_block = DOUBLY_INDIRECT_MAX + Block::new(42);
    let physical_block = Some(Block::new(100));
    converter.write_physical(&mut inode, inode_block, physical_block)?;

    // Then expect the inode was updated
    let mut wanted_inode = Inode::new(Ino::new(0));
    let triply_indirect_block = Block::new(0);
    wanted_inode.triply_indirect_block = Some(triply_indirect_block);
    wanted_inode.size = BLOCK_SIZE * Byte::new(3);
    assert_eq!(wanted_inode, inode);

    // And expect the indirect converter was updated
    let mut wanted_table = HashMap::new();
    let triply_indirect_index = BlockListIndex::new(0);
    let doubly_indirect_index = BlockListIndex::new(0);
    let doubly_indirect_block = Block::new(1);
    let singly_indirect_block = Block::new(2);
    let singly_indirect_index = BlockListIndex::new(42);
    wanted_table.insert(
        (triply_indirect_block, triply_indirect_index),
        Some(doubly_indirect_block),
    );
    wanted_table.insert(
        (doubly_indirect_block, doubly_indirect_index),
        Some(singly_indirect_block),
    );
    wanted_table.insert(
        (singly_indirect_block, singly_indirect_index),
        physical_block,
    );
    assert_eq!(wanted_table, converter.indirect_manager.table);

    Ok(())
}

#[test]
fn test_write_triply_indirect_allocate_two() -> Result<()> {
    let mut converter = PhysicalWriter::default();
    let mut inode = Inode::new(Ino::new(0));

    // Given *only* the triply indirect block has been prepared for the inode
    let triply_indirect_block = alloc(&mut converter)?;
    inode.triply_indirect_block = Some(triply_indirect_block);
    inode.size = BLOCK_SIZE;

    // When we try to set an inode block that is outside of the doubly-indirect
    // range
    let inode_block = DOUBLY_INDIRECT_MAX + Block::new(42);
    let physical_block = Some(Block::new(100));
    converter.write_physical(&mut inode, inode_block, physical_block)?;

    // Then expect the inode was updated
    let mut wanted_inode = Inode::new(Ino::new(0));
    wanted_inode.triply_indirect_block = Some(triply_indirect_block);
    wanted_inode.size = BLOCK_SIZE * Byte::new(3);
    assert_eq!(wanted_inode, inode);

    // And expect the indirect converter was updated
    let mut wanted_table = HashMap::new();
    let triply_indirect_index = BlockListIndex::new(0);
    let doubly_indirect_index = BlockListIndex::new(0);
    let doubly_indirect_block = Block::new(1);
    let singly_indirect_block = Block::new(2);
    let singly_indirect_index = BlockListIndex::new(42);
    wanted_table.insert(
        (triply_indirect_block, triply_indirect_index),
        Some(doubly_indirect_block),
    );
    wanted_table.insert(
        (doubly_indirect_block, doubly_indirect_index),
        Some(singly_indirect_block),
    );
    wanted_table.insert(
        (singly_indirect_block, singly_indirect_index),
        physical_block,
    );
    assert_eq!(wanted_table, converter.indirect_manager.table);

    Ok(())
}

#[test]
fn test_write_triply_indirect_allocate_one() -> Result<()> {
    let mut converter = PhysicalWriter::default();
    let mut inode = Inode::new(Ino::new(0));

    // Given *only* the triply and doubly indirect blocks have been prepared
    // for the inode
    let triply_indirect_index = BlockListIndex::new(0);
    let triply_indirect_block = alloc(&mut converter)?;
    let doubly_indirect_block = alloc(&mut converter)?;
    inode.triply_indirect_block = Some(triply_indirect_block);
    inode.size = BLOCK_SIZE * Byte::new(2);
    converter.indirect_manager.table.insert(
        (triply_indirect_block, triply_indirect_index),
        Some(doubly_indirect_block),
    );

    // When we try to set an inode block that is outside of the doubly-indirect
    // range
    let inode_block = DOUBLY_INDIRECT_MAX + Block::new(42);
    let physical_block = Some(Block::new(100));
    converter.write_physical(&mut inode, inode_block, physical_block)?;

    // Then expect the inode was updated
    let mut wanted_inode = Inode::new(Ino::new(0));
    wanted_inode.triply_indirect_block = Some(triply_indirect_block);
    wanted_inode.size = BLOCK_SIZE * Byte::new(3);
    assert_eq!(wanted_inode, inode);

    // And expect the indirect converter was updated
    let mut wanted_table = HashMap::new();
    let doubly_indirect_index = BlockListIndex::new(0);
    let doubly_indirect_block = Block::new(1);
    let singly_indirect_block = Block::new(2);
    let singly_indirect_index = BlockListIndex::new(42);
    wanted_table.insert(
        (triply_indirect_block, triply_indirect_index),
        Some(doubly_indirect_block),
    );
    wanted_table.insert(
        (doubly_indirect_block, doubly_indirect_index),
        Some(singly_indirect_block),
    );
    wanted_table.insert(
        (singly_indirect_block, singly_indirect_index),
        physical_block,
    );
    assert_eq!(wanted_table, converter.indirect_manager.table);

    Ok(())
}

#[test]
fn test_write_triply_indirect_allocate_none() -> Result<()> {
    let mut converter = PhysicalWriter::default();
    let mut inode = Inode::new(Ino::new(0));

    // Given the triply, doubly, and singly indirect blocks have been prepared
    // for the inode
    let triply_indirect_index = BlockListIndex::new(0);
    let doubly_indirect_index = BlockListIndex::new(0);
    let triply_indirect_block = alloc(&mut converter)?;
    let doubly_indirect_block = alloc(&mut converter)?;
    let singly_indirect_block = alloc(&mut converter)?;
    inode.triply_indirect_block = Some(triply_indirect_block);
    inode.size = BLOCK_SIZE * Byte::new(3);
    converter.indirect_manager.table.insert(
        (triply_indirect_block, triply_indirect_index),
        Some(doubly_indirect_block),
    );
    converter.indirect_manager.table.insert(
        (doubly_indirect_block, doubly_indirect_index),
        Some(singly_indirect_block),
    );

    // When we try to set an inode block that is outside of the doubly-indirect
    // range
    let inode_block = DOUBLY_INDIRECT_MAX + Block::new(42);
    let physical_block = Some(Block::new(100));
    converter.write_physical(&mut inode, inode_block, physical_block)?;

    // Then expect the inode was updated
    let mut wanted_inode = Inode::new(Ino::new(0));
    wanted_inode.triply_indirect_block = Some(triply_indirect_block);
    wanted_inode.size = BLOCK_SIZE * Byte::new(3);
    assert_eq!(wanted_inode, inode);

    // And expect the indirect converter was updated
    let mut wanted_table = HashMap::new();
    let doubly_indirect_index = BlockListIndex::new(0);
    let doubly_indirect_block = Block::new(1);
    let singly_indirect_block = Block::new(2);
    let singly_indirect_index = BlockListIndex::new(42);
    wanted_table.insert(
        (triply_indirect_block, triply_indirect_index),
        Some(doubly_indirect_block),
    );
    wanted_table.insert(
        (doubly_indirect_block, doubly_indirect_index),
        Some(singly_indirect_block),
    );
    wanted_table.insert(
        (singly_indirect_block, singly_indirect_index),
        physical_block,
    );
    assert_eq!(wanted_table, converter.indirect_manager.table);

    Ok(())
}

#[test]
fn test_write_out_of_range() -> Result<()> {
    let mut converter = PhysicalWriter::default();
    let mut inode = Inode::new(Ino::new(0));

    // When we try to set an inode block that is outside of even the
    // triply-indirect range
    let inode_block = TRIPLY_INDIRECT_MAX + Block::new(42);
    let physical_block = Some(Block::new(100));

    // Then expect an out-of-range error
    match converter.write_physical(&mut inode, inode_block, physical_block) {
        Err(IndirectionError::OutOfRange) => Ok(()),
        Err(e) => panic!("unexpected error: {0}", e),
        Ok(_) => panic!("unexpected success"),
    }
}

impl<'a> Default for PhysicalWriter<FakeIndirectManager, HashMap<Ino, Inode>> {
    fn default() -> PhysicalWriter<FakeIndirectManager, HashMap<Ino, Inode>> {
        PhysicalWriter::from([])
    }
}

impl<H> From<H> for PhysicalWriter<FakeIndirectManager, HashMap<Ino, Inode>>
where
    FakeIndirectManager: From<H>,
{
    fn from(h: H) -> PhysicalWriter<FakeIndirectManager, HashMap<Ino, Inode>> {
        PhysicalWriter {
            indirect_manager: FakeIndirectManager::from(h),
            inode_store: HashMap::new(),
            block_allocator: Bitmap::new(Byte::new(10)),
        }
    }
}

struct FakeIndirectManager {
    table: HashMap<(Block, BlockListIndex), Option<Block>>,
}

impl FakeIndirectManager {
    fn new() -> FakeIndirectManager {
        FakeIndirectManager {
            table: HashMap::new(),
        }
    }
}

impl<H> From<H> for FakeIndirectManager
where
    HashMap<(Block, BlockListIndex), Option<Block>>: From<H>,
{
    fn from(table: H) -> FakeIndirectManager {
        FakeIndirectManager {
            table: HashMap::from(table),
        }
    }
}

impl IndirectReader for FakeIndirectManager {
    fn read_indirect(
        &self,
        indirect: Block,
        index: BlockListIndex,
    ) -> std::io::Result<Option<Block>> {
        match self.table.get(&(indirect, index)) {
            Some(target) => Ok(*target),
            None => Ok(None),
        }
    }
}

impl IndirectWriter for FakeIndirectManager {
    fn write_indirect(
        &mut self,
        indirect: Block,
        index: BlockListIndex,
        target: Option<Block>,
    ) -> std::io::Result<()> {
        self.table.insert((indirect, index), target);
        Ok(())
    }
}

impl IndirectManager for FakeIndirectManager {}
