use super::{ReadAt, WriteAt};
use crate::byte::Byte;
use std::{io::Result, ops::Add};

pub struct OffsetVolume<B> {
    backend: B,
    offset: Byte,
}

impl<B> OffsetVolume<B> {
    pub fn new(backend: B, offset: Byte) -> OffsetVolume<B> {
        OffsetVolume { backend, offset }
    }
}

impl<B: ReadAt> ReadAt for OffsetVolume<B> {
    fn read_at(&self, offset: Byte, buf: &mut [u8]) -> Result<()> {
        self.backend.read_at(offset + self.offset, buf)
    }
}

impl<B: WriteAt> WriteAt for OffsetVolume<B> {
    fn write_at(&mut self, offset: Byte, buf: &[u8]) -> Result<()> {
        self.backend.write_at(offset + self.offset, buf)
    }
}
