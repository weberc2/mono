use super::manager::*;
use super::Result;
use crate::block::{Block, BLOCK_SIZE};
use crate::byte::Byte;
use crate::indirectmanager::HashMap;
use crate::inode::{Ino, Inode};
use std::io::Cursor;

#[test]
fn test_write() -> Result<()> {
    let volume = vec![0 as u8; 1024];
    let mut manager = Manager::new(HashMap::new().into(), volume);
    let inode = Inode {
        ino: Ino::new(0),
        size: BLOCK_SIZE,
        direct_blocks: [
            Some(Block::new(0)),
            None,
            None,
            None,
            None,
            None,
            None,
            None,
            None,
            None,
            None,
            None,
        ],
        singly_indirect_block: None,
        doubly_indirect_block: None,
        triply_indirect_block: None,
    };
    manager.write_inode_block(&inode, Block::new(0), Byte::new(0), &[1, 1, 1, 1, 1])?;

    let mut buf = [0; BLOCK_SIZE.to_usize()];
    manager.read_inode_block(&inode, Block::new(0), Byte::new(0), &mut buf)?;

    let mut wanted = [0 as u8; BLOCK_SIZE.to_usize()];
    wanted[..5].copy_from_slice(&[1, 1, 1, 1, 1]);

    assert_eq!(wanted, buf);

    Ok(())
}
