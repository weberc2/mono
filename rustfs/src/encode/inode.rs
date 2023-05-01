use crate::block::*;
use crate::byte::*;
use crate::inode::{Ino, Inode};

pub const INODE_SIZE: usize = 1024;
const SINGLY_INDIRECT_START: usize =
    DIRECT_BLOCKS_COUNT * BLOCK_POINTER_SIZE.to_usize() + BYTE_POINTER_SIZE.to_usize();
const DOUBLY_INDIRECT_START: usize = SINGLY_INDIRECT_START + BLOCK_POINTER_SIZE.to_usize();
const TRIPLY_INDIRECT_START: usize = DOUBLY_INDIRECT_START + BLOCK_POINTER_SIZE.to_usize();

pub fn encode_inode(inode: &Inode, buf: &mut [u8; INODE_SIZE]) {
    use byteorder::{ByteOrder, LittleEndian};
    LittleEndian::write_u64(buf, u64::from(inode.size));
    for (i, block) in inode.direct_blocks.iter().enumerate() {
        LittleEndian::write_u64(
            &mut buf[i * usize::from(BLOCK_POINTER_SIZE)..],
            Block::encode(*block),
        )
    }
    LittleEndian::write_u64(
        &mut buf[SINGLY_INDIRECT_START..],
        Block::encode(inode.singly_indirect_block),
    );

    LittleEndian::write_u64(
        &mut buf[DOUBLY_INDIRECT_START..],
        Block::encode(inode.doubly_indirect_block),
    );

    LittleEndian::write_u64(
        &mut buf[TRIPLY_INDIRECT_START..],
        Block::encode(inode.triply_indirect_block),
    );
}

pub fn decode_inode(buf: &[u8; INODE_SIZE], ino: Ino) -> Inode {
    use byteorder::{ByteOrder, LittleEndian};
    let mut inode = Inode::new(ino);
    inode.size = Byte::new(LittleEndian::read_u64(buf));
    for (i, block) in inode.direct_blocks.iter_mut().enumerate() {
        *block = Block::decode(LittleEndian::read_u64(
            &buf[i * usize::from(BLOCK_POINTER_SIZE)..],
        ));
    }
    inode.singly_indirect_block =
        Block::decode(LittleEndian::read_u64(&buf[SINGLY_INDIRECT_START..]));
    inode.doubly_indirect_block =
        Block::decode(LittleEndian::read_u64(&buf[DOUBLY_INDIRECT_START..]));
    inode.triply_indirect_block =
        Block::decode(LittleEndian::read_u64(&buf[TRIPLY_INDIRECT_START..]));
    inode
}
