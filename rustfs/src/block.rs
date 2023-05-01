use crate::byte::Byte;

pub const DIRECT_BLOCKS_COUNT: usize = 12;
pub const BLOCK_SIZE: Byte = Byte::new(1024);
pub const BLOCK_POINTER_SIZE: Byte = Byte::new(8);
pub const BLOCK_POINTERS_PER_BLOCK: u64 = BLOCK_SIZE.to_u64() / BLOCK_POINTER_SIZE.to_u64();
pub const DIRECT_BLOCKS_MAX: Block = Block((DIRECT_BLOCKS_COUNT - 1) as u64);
pub const SINGLY_INDIRECT_COUNT: Block = Block(BLOCK_POINTERS_PER_BLOCK as u64);
pub const SINGLY_INDIRECT_MAX: Block = Block(SINGLY_INDIRECT_COUNT.0 + DIRECT_BLOCKS_MAX.0);
pub const DOUBLY_INDIRECT_COUNT: Block = Block(SINGLY_INDIRECT_COUNT.0 * BLOCK_POINTERS_PER_BLOCK);
pub const DOUBLY_INDIRECT_MAX: Block = Block(DOUBLY_INDIRECT_COUNT.0 + SINGLY_INDIRECT_MAX.0);
pub const TRIPLY_INDIRECT_COUNT: Block = Block(DOUBLY_INDIRECT_COUNT.0 * BLOCK_POINTERS_PER_BLOCK);
pub const TRIPLY_INDIRECT_MAX: Block = Block(TRIPLY_INDIRECT_COUNT.0 + DOUBLY_INDIRECT_MAX.0);

#[derive(Debug, Clone, Copy, PartialEq, PartialOrd, Eq, Hash)]
pub struct Block(u64);

impl Block {
    pub const fn new(value: u64) -> Block {
        Block(value)
    }

    pub const fn to_usize(self) -> usize {
        self.0 as usize
    }

    pub const fn to_u64(self) -> u64 {
        self.0
    }

    pub const fn encode(block: Option<Block>) -> u64 {
        match block {
            None => 0,
            Some(block) => block.to_u64() + 1,
        }
    }

    pub const fn decode(u: u64) -> Option<Self> {
        match u {
            0 => None,
            block => Some(Block::new(block - 1)),
        }
    }
}

impl std::fmt::Display for Block {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.0)
    }
}

impl From<u64> for Block {
    fn from(u: u64) -> Block {
        Block(u)
    }
}

impl From<Block> for u64 {
    fn from(b: Block) -> u64 {
        b.0
    }
}

impl std::ops::Add for Block {
    type Output = Self;

    fn add(self, other: Self) -> Self::Output {
        Self(self.0 + other.0)
    }
}

impl std::ops::Sub for Block {
    type Output = Self;

    fn sub(self, other: Self) -> Self::Output {
        Self(self.0 - other.0)
    }
}

impl std::ops::Mul for Block {
    type Output = Self;

    fn mul(self, other: Self) -> Self::Output {
        Self(self.0 * other.0)
    }
}

impl std::ops::Div for Block {
    type Output = Self;

    fn div(self, other: Self) -> Self::Output {
        Self(self.0 / other.0)
    }
}

impl std::ops::Rem for Block {
    type Output = Self;

    fn rem(self, other: Self) -> Self::Output {
        Self(self.0 % other.0)
    }
}

#[derive(Debug, Clone, Copy, PartialEq, PartialOrd, Eq, Hash)]
pub struct BlockListIndex(usize);

impl BlockListIndex {
    pub fn new(value: usize) -> BlockListIndex {
        BlockListIndex(value)
    }

    pub fn set_direct_block(
        self,
        direct_blocks: &mut [Option<Block>; DIRECT_BLOCKS_COUNT],
        block: Option<Block>,
    ) {
        direct_blocks[self.0] = block;
    }

    pub fn get_direct_block(
        self,
        direct_blocks: &[Option<Block>; DIRECT_BLOCKS_COUNT],
    ) -> Option<Block> {
        direct_blocks[self.0]
    }
}

impl From<BlockListIndex> for u64 {
    fn from(index: BlockListIndex) -> u64 {
        index.0 as u64
    }
}

impl std::ops::Add for BlockListIndex {
    type Output = Self;

    fn add(self, other: Self) -> Self::Output {
        Self(self.0 + other.0)
    }
}

impl std::ops::Sub for BlockListIndex {
    type Output = Self;

    fn sub(self, other: Self) -> Self::Output {
        Self(self.0 - other.0)
    }
}

impl std::ops::Mul for BlockListIndex {
    type Output = Self;

    fn mul(self, other: Self) -> Self::Output {
        Self(self.0 * other.0)
    }
}

impl std::ops::Div for BlockListIndex {
    type Output = Self;

    fn div(self, other: Self) -> Self::Output {
        Self(self.0 / other.0)
    }
}

impl std::ops::Rem for BlockListIndex {
    type Output = Self;

    fn rem(self, other: Self) -> Self::Output {
        Self(self.0 % other.0)
    }
}

#[derive(Debug, PartialEq)]
pub enum Indirection {
    Direct {
        direct_index: BlockListIndex,
    },
    SinglyIndirect {
        singly_indirect_index: BlockListIndex,
    },
    DoublyIndirect {
        doubly_indirect_index: BlockListIndex,
        singly_indirect_index: BlockListIndex,
    },
    TriplyIndirect {
        triply_indirect_index: BlockListIndex,
        doubly_indirect_index: BlockListIndex,
        singly_indirect_index: BlockListIndex,
    },
    OutOfRange,
}

// singly
// |____
// | | |

// doubly
// |______________
// |____  |____  |____
// | | |  | | |  | | |

// triply
// |____________________________________________
// |______________       |______________       |______________
// |____  |____  |____   |____  |____  |____   |____  |____  |____
// | | |  | | |  | | |   | | |  | | |  | | |   | | |  | | |  | | |

impl From<Block> for Indirection {
    fn from(indirect_block: Block) -> Indirection {
        if indirect_block <= DIRECT_BLOCKS_MAX {
            Indirection::Direct {
                direct_index: BlockListIndex(indirect_block.0 as usize),
            }
        } else if indirect_block <= SINGLY_INDIRECT_MAX {
            Indirection::SinglyIndirect {
                singly_indirect_index: BlockListIndex(
                    (indirect_block - DIRECT_BLOCKS_MAX).0 as usize,
                ),
            }
        } else if indirect_block <= DOUBLY_INDIRECT_MAX {
            let base = indirect_block - SINGLY_INDIRECT_MAX;
            Indirection::DoublyIndirect {
                doubly_indirect_index: BlockListIndex((base / SINGLY_INDIRECT_COUNT).0 as usize),
                singly_indirect_index: BlockListIndex((base % SINGLY_INDIRECT_COUNT).0 as usize),
            }
        } else if indirect_block <= TRIPLY_INDIRECT_MAX {
            let base = indirect_block - DOUBLY_INDIRECT_MAX;
            Indirection::TriplyIndirect {
                triply_indirect_index: BlockListIndex((base / DOUBLY_INDIRECT_COUNT).0 as usize),
                doubly_indirect_index: BlockListIndex(
                    ((base % DOUBLY_INDIRECT_COUNT) / SINGLY_INDIRECT_COUNT).0 as usize,
                ),
                singly_indirect_index: BlockListIndex(
                    ((base % DOUBLY_INDIRECT_COUNT) % SINGLY_INDIRECT_COUNT).0 as usize,
                ),
            }
        } else {
            Indirection::OutOfRange
        }
    }
}

#[cfg(test)]
mod tests {
    use crate::block::*;
    #[test]
    fn test_indirection_from_block_doubly_indirect() {
        assert_eq!(
            Indirection::DoublyIndirect {
                doubly_indirect_index: BlockListIndex(0),
                singly_indirect_index: BlockListIndex(42),
            },
            Indirection::from(SINGLY_INDIRECT_MAX + Block(42)),
        );
    }

    #[test]
    fn test_indirection_from_block_triply_indirect() {
        assert_eq!(
            Indirection::TriplyIndirect {
                triply_indirect_index: BlockListIndex(0),
                doubly_indirect_index: BlockListIndex(0),
                singly_indirect_index: BlockListIndex(42)
            },
            Indirection::from(DOUBLY_INDIRECT_MAX + Block(42)),
        );
    }
}
