use crate::block::*;
use crate::byte::*;

#[derive(Clone, Copy, PartialEq, Eq, Debug, Hash)]
pub struct Ino(u64);

impl Ino {
    pub fn new(value: u64) -> Ino {
        Ino(value)
    }
}

impl std::fmt::Display for Ino {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{0}", self.0)
    }
}

impl From<Ino> for Byte {
    fn from(ino: Ino) -> Byte {
        Byte::new(ino.0)
    }
}

#[derive(Clone, PartialEq, Eq, Debug)]
pub struct Inode {
    pub ino: Ino,
    pub size: Byte,
    pub direct_blocks: [Option<Block>; DIRECT_BLOCKS_COUNT],
    pub singly_indirect_block: Option<Block>,
    pub doubly_indirect_block: Option<Block>,
    pub triply_indirect_block: Option<Block>,
}

impl Inode {
    pub fn new(ino: Ino) -> Inode {
        Inode {
            ino: ino,
            size: Byte::new(0),
            direct_blocks: [
                None, None, None, None, None, None, None, None, None, None, None, None,
            ],
            singly_indirect_block: None,
            doubly_indirect_block: None,
            triply_indirect_block: None,
        }
    }

    pub fn set_direct_block(&mut self, index: BlockListIndex, block: Option<Block>) {
        index.set_direct_block(&mut self.direct_blocks, block)
    }
}
