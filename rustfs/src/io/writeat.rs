use crate::byte::Byte;
use std::io::{Error, ErrorKind, Result, Seek, SeekFrom, Write};
use std::ops::DerefMut;

pub trait WriteAt {
    fn write_at(&mut self, offset: Byte, b: &[u8]) -> Result<()>;
}

impl<'a, D: DerefMut<Target = [u8]>> WriteAt for D {
    fn write_at(&mut self, offset: Byte, b: &[u8]) -> Result<()> {
        let offset = offset.to_usize();
        let size = std::cmp::min(self.len() - offset, b.len());
        if size < 1 {
            Err(Error::from(ErrorKind::UnexpectedEof))
        } else {
            self[..size].copy_from_slice(&b[..size]);
            Ok(())
        }
    }
}
