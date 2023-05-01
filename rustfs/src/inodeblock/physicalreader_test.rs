use super::physicalreader::PhysicalReader;
use super::Result;
use crate::block::{
    Block, BlockListIndex, BLOCK_POINTERS_PER_BLOCK, DIRECT_BLOCKS_COUNT, DIRECT_BLOCKS_MAX,
    DOUBLY_INDIRECT_MAX,
};
use crate::byte::Byte;
use crate::indirectmanager::HashMap;
use crate::inode::{Ino, Inode};

impl Default for PhysicalReader<HashMap> {
    fn default() -> Self {
        Self::new(HashMap::default())
    }
}

impl<T> From<T> for PhysicalReader<HashMap>
where
    HashMap: From<T>,
{
    fn from(value: T) -> Self {
        Self::new(HashMap::from(value))
    }
}

#[test]
fn test_read_direct() -> Result<()> {
    let mut physical_reader = PhysicalReader::default();

    let physical_block = physical_reader.read_physical(
        &Inode {
            ino: Ino::new(0),
            size: Byte::new(0),
            direct_blocks: [0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22].map(|x| Some(Block::new(x))),
            singly_indirect_block: None,
            doubly_indirect_block: None,
            triply_indirect_block: None,
        },
        Block::new(1),
    )?;

    assert_eq!(physical_block, Some(Block::new(2)));
    Ok(())
}

#[test]
fn test_read_singly_indirect() -> Result<()> {
    let mut physical_reader = PhysicalReader::from([(
        (Block::new(1), BlockListIndex::new(42)),
        Some(Block::new(1000)),
    )]);

    let physical_block = physical_reader.read_physical(
        &Inode {
            ino: Ino::new(0),
            size: Byte::new(0),
            direct_blocks: [0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22].map(|x| Some(Block::new(x))),
            singly_indirect_block: Some(Block::new(1)),
            doubly_indirect_block: None,
            triply_indirect_block: None,
        },
        DIRECT_BLOCKS_MAX + Block::new(42),
    )?;

    assert_eq!(physical_block, Some(Block::new(1000)));
    Ok(())
}

#[test]
fn test_read_singly_indirect_no_singly_indirect_block() -> Result<()> {
    let mut physical_reader = PhysicalReader::from([(
        (Block::new(1), BlockListIndex::new(42)),
        Some(Block::new(1000)),
    )]);

    let physical_block = physical_reader.read_physical(
        &Inode {
            ino: Ino::new(0),
            size: Byte::new(0),
            direct_blocks: [0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22].map(|x| Some(Block::new(x))),
            singly_indirect_block: None,
            doubly_indirect_block: None,
            triply_indirect_block: None,
        },
        Block::new((DIRECT_BLOCKS_COUNT + 42) as u64),
    )?;

    assert_eq!(physical_block, None);
    Ok(())
}

#[test]
fn test_read_doubly_indirect() -> Result<()> {
    let mut physical_reader = PhysicalReader::from([
        (
            (Block::new(10), BlockListIndex::new(42)),
            Some(Block::new(1000)),
        ),
        (
            (Block::new(1), BlockListIndex::new(0)),
            Some(Block::new(10)),
        ),
    ]);

    let physical_block = physical_reader.read_physical(
        &Inode {
            ino: Ino::new(0),
            size: Byte::new(0),
            direct_blocks: [0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22].map(|x| Some(Block::new(x))),
            singly_indirect_block: None,
            doubly_indirect_block: Some(Block::new(1)),
            triply_indirect_block: None,
        },
        Block::new((DIRECT_BLOCKS_COUNT + 41) as u64 + BLOCK_POINTERS_PER_BLOCK),
    )?;

    assert_eq!(physical_block, Some(Block::new(1000)));
    Ok(())
}

#[test]
fn test_read_triply_indirect() -> Result<()> {
    let mut physical_reader = PhysicalReader::from([
        (
            (Block::new(1000), BlockListIndex::new(42)),
            Some(Block::new(1001)),
        ),
        (
            (Block::new(10), BlockListIndex::new(0)),
            Some(Block::new(1000)),
        ),
        (
            (Block::new(1), BlockListIndex::new(0)),
            Some(Block::new(10)),
        ),
    ]);

    let physical_block = physical_reader.read_physical(
        &Inode {
            ino: Ino::new(0),
            size: Byte::new(0),
            direct_blocks: [0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22].map(|x| Some(Block::new(x))),
            singly_indirect_block: None,
            doubly_indirect_block: None,
            triply_indirect_block: Some(Block::new(1)),
        },
        DOUBLY_INDIRECT_MAX + Block::new(42),
    )?;

    assert_eq!(physical_block, Some(Block::new(1001)));
    Ok(())
}
