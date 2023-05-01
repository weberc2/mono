use crate::byte::Byte;
use std::io::{Error, ErrorKind, Read, Result, Seek, SeekFrom};
use std::ops::Deref;

pub trait ReadAt {
    fn read_at(&self, offset: Byte, b: &mut [u8]) -> Result<()>;
}

impl<D: Deref<Target = [u8]>> ReadAt for D {
    fn read_at(&self, offset: Byte, b: &mut [u8]) -> Result<()> {
        let offset = offset.to_usize();
        let size = std::cmp::min(self.len() - offset, b.len());
        if size < 1 {
            Err(Error::from(ErrorKind::UnexpectedEof))
        } else {
            b.copy_from_slice(&self[offset..offset + size]);
            Ok(())
        }
    }
}

// fn read_at(&'a self, offset: Byte, b: &mut [u8]) -> Result<()> {
//     let slice: &[u8] = &[u8]::from(self);
//     let offset = offset.to_usize();
//     let size = std::cmp::min(slice.len() - offset, b.len());
//     if size < 1 {
//         Err(Error::from(ErrorKind::UnexpectedEof))
//     } else {
//         b.copy_from_slice(&slice[offset..offset + size]);
//         Ok(())
//     }
// }
