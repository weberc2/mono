use crate::byte::Byte;
use crate::encode::{decode_inode, encode_inode, INODE_SIZE};
use crate::inode::{Ino, Inode};
use crate::io::{ReadAt, WriteAt};
use std::io::Result;

pub fn write_inode<W: WriteAt>(w: &mut W, inode: &Inode) -> Result<()> {
    let mut buf: [u8; INODE_SIZE] = [0; INODE_SIZE];
    encode_inode(inode, &mut buf);
    w.write_at(Byte::from(inode.ino) * Byte::new(INODE_SIZE as u64), &buf)
}

pub fn read_inode<R: ReadAt>(r: &mut R, ino: Ino) -> Result<Inode> {
    let mut buf: [u8; INODE_SIZE] = [0; INODE_SIZE];
    r.read_at(Byte::from(ino) * Byte::new(INODE_SIZE as u64), &mut buf)?;
    Ok(decode_inode(&buf, ino))
}
