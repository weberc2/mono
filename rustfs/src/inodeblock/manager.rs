use super::physicalreader::PhysicalReader;
use super::{IndirectionError, Result};
use crate::block::{Block, BLOCK_SIZE};
use crate::byte::Byte;
use crate::indirectmanager::IndirectReader;
use crate::inode::Inode;
use crate::io::{ReadAt, WriteAt};
use std::io::{Error, ErrorKind};

pub trait Reader {
    fn read_inode_block(
        &self,
        inode: &Inode,
        inode_block: Block,
        offset: Byte,
        buf: &mut [u8],
    ) -> Result<()>;
}

impl<R: IndirectReader, V: ReadAt> Reader for Manager<R, V> {
    fn read_inode_block(
        &self,
        inode: &Inode,
        inode_block: Block,
        offset: Byte,
        buf: &mut [u8],
    ) -> Result<()> {
        assert!(offset + Byte::from(buf.len()) <= BLOCK_SIZE);
        match self.physical_reader.read_physical(inode, inode_block)? {
            Some(physical_block) => self
                .volume
                .read_at(Byte::from((physical_block, offset)), buf)
                .map_err(IndirectionError::from),
            None => Err(IndirectionError::from(Error::from(
                ErrorKind::UnexpectedEof,
            ))),
        }
    }
}

impl<R: Reader> Reader for &mut R {
    fn read_inode_block(
        &self,
        inode: &Inode,
        inode_block: Block,
        offset: Byte,
        buf: &mut [u8],
    ) -> Result<()> {
        (**self).read_inode_block(inode, inode_block, offset, buf)
    }
}

pub trait Writer {
    fn write_inode_block(
        &mut self,
        inode: &Inode,
        inode_block: Block,
        offset: Byte,
        buf: &[u8],
    ) -> Result<()>;
}

impl<W: Writer> Writer for &mut W {
    fn write_inode_block(
        &mut self,
        inode: &Inode,
        inode_block: Block,
        offset: Byte,
        buf: &[u8],
    ) -> Result<()> {
        (**self).write_inode_block(inode, inode_block, offset, buf)
    }
}

impl<R: IndirectReader, V: WriteAt> Writer for Manager<R, V> {
    fn write_inode_block(
        &mut self,
        inode: &Inode,
        inode_block: Block,
        offset: Byte,
        buf: &[u8],
    ) -> Result<()> {
        crate::annotate!(
            match self.physical_reader.read_physical(inode, inode_block)? {
                Some(physical_block) => self
                    .volume
                    .write_at(Byte::from((physical_block, offset)), buf)
                    .map_err(IndirectionError::from),
                None => Err(IndirectionError::from(Error::from(
                    ErrorKind::UnexpectedEof,
                ))),
            },
            "writing `{}` bytes to inode `{}` block `{}` at offset `{}`",
            buf.len(),
            inode.ino,
            inode_block,
            offset,
        )
    }
}

/// Converts a `(Block, Byte)` (a relative byte-offset within a block) into an
/// absolute byte offset.
impl From<(Block, Byte)> for Byte {
    fn from((block, byte): (Block, Byte)) -> Byte {
        Byte::from(u64::from(block)) * BLOCK_SIZE + byte
    }
}

pub struct Manager<R: IndirectReader, V> {
    physical_reader: PhysicalReader<R>,
    volume: V,
}

impl<R: IndirectReader, V> Manager<R, V> {
    pub fn new(physical_reader: PhysicalReader<R>, volume: V) -> Manager<R, V> {
        Manager {
            physical_reader,
            volume,
        }
    }
}
