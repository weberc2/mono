use crate::byte::Byte;
use std::marker::PhantomData;

pub struct Bitmap<T> {
    bytes: Vec<u8>,
    r#type: PhantomData<T>,
}

impl<T> Bitmap<T>
where
    T: From<u64>,
    u64: From<T>,
{
    pub fn new(size: Byte) -> Bitmap<T> {
        Bitmap {
            bytes: vec![0; size.to_usize()],
            r#type: PhantomData {},
        }
    }

    fn first_zero(bytes: &[u8]) -> Option<(usize, u8)> {
        for (i, byte) in bytes.iter().enumerate() {
            if let Some(bit) = Bits::first_zero(*byte) {
                return Some((i, bit));
            }
        }
        None
    }

    pub fn allocate(&mut self) -> Option<T> {
        Bitmap::first_zero(&self.bytes).map(|(byte, bit)| {
            self.bytes[byte] = Bits::set_high(self.bytes[byte], bit);
            T::from(byte as u64 * 8 + bit as u64)
        })
    }

    pub fn free(&mut self, value: T) {
        let value = u64::from(value);
        let byte = value as usize / 8;
        let bit = value as u8 % 8;
        self.bytes[byte] = Bits::set_low(self.bytes[byte], bit);
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_allocate() {
        let mut bitmap: Bitmap<u64> = Bitmap::new(Byte::new(1));
        assert_eq!(Some(0), bitmap.allocate());
        assert_eq!(Some(1), bitmap.allocate());
        assert_eq!(Some(2), bitmap.allocate());
        assert_eq!(Some(3), bitmap.allocate());
        assert_eq!(Some(4), bitmap.allocate());
        assert_eq!(Some(5), bitmap.allocate());
        assert_eq!(Some(6), bitmap.allocate());
        assert_eq!(Some(7), bitmap.allocate());
        assert_eq!(None, bitmap.allocate());

        bitmap.free(4);
        assert_eq!(Some(4), bitmap.allocate());
    }
}

struct Bits;

impl Bits {
    fn is_zero(byte: u8, bit: u8) -> bool {
        byte & (0b1000_0000 >> bit) == 0
    }

    fn set_high(byte: u8, bit: u8) -> u8 {
        byte | 0b10000000 >> bit
    }

    fn set_low(byte: u8, bit: u8) -> u8 {
        byte & !(0b1000_0000 >> bit)
    }

    fn first_zero(byte: u8) -> Option<u8> {
        for bit in 0..8 {
            if Bits::is_zero(byte, bit) {
                return Some(bit);
            }
        }
        None
    }
}

#[cfg(test)]
mod bitstests {
    use super::*;

    #[test]
    fn test_set_high() {
        assert_eq!(0b1000_0000, Bits::set_high(0, 0));
        assert_eq!(0b0100_0000, Bits::set_high(0, 1));
        assert_eq!(0b0010_0000, Bits::set_high(0, 2));
        assert_eq!(0b0001_0000, Bits::set_high(0, 3));
        assert_eq!(0b0000_1000, Bits::set_high(0, 4));
        assert_eq!(0b0000_0100, Bits::set_high(0, 5));
        assert_eq!(0b0000_0010, Bits::set_high(0, 6));
        assert_eq!(0b0000_0001, Bits::set_high(0, 7));
    }

    #[test]
    fn test_set_low() {
        assert_eq!(0b0111_1111, Bits::set_low(0xFF, 0));
        assert_eq!(0b1011_1111, Bits::set_low(0xFF, 1));
        assert_eq!(0b1101_1111, Bits::set_low(0xFF, 2));
        assert_eq!(0b1110_1111, Bits::set_low(0xFF, 3));
        assert_eq!(0b1111_0111, Bits::set_low(0xFF, 4));
        assert_eq!(0b1111_1011, Bits::set_low(0xFF, 5));
        assert_eq!(0b1111_1101, Bits::set_low(0xFF, 6));
        assert_eq!(0b1111_1110, Bits::set_low(0xFF, 7));
    }

    #[test]
    fn test_is_zero() {
        assert!(Bits::is_zero(0b0000_0000, 0));
        assert!(Bits::is_zero(0b0000_0000, 1));
        assert!(Bits::is_zero(0b0000_0000, 2));
        assert!(Bits::is_zero(0b0000_0000, 3));
        assert!(Bits::is_zero(0b0000_0000, 4));
        assert!(Bits::is_zero(0b0000_0000, 5));
        assert!(Bits::is_zero(0b0000_0000, 6));
        assert!(Bits::is_zero(0b0000_0000, 7));

        assert!(!Bits::is_zero(0b1111_1111, 0));
        assert!(!Bits::is_zero(0b1111_1111, 1));
        assert!(!Bits::is_zero(0b1111_1111, 2));
        assert!(!Bits::is_zero(0b1111_1111, 3));
        assert!(!Bits::is_zero(0b1111_1111, 4));
        assert!(!Bits::is_zero(0b1111_1111, 5));
        assert!(!Bits::is_zero(0b1111_1111, 6));
        assert!(!Bits::is_zero(0b1111_1111, 7));
    }

    #[test]
    fn test_first_zero() {
        assert_eq!(Some(0), Bits::first_zero(0b0000_0000));
        assert_eq!(Some(1), Bits::first_zero(0b1000_0000));
    }
}
