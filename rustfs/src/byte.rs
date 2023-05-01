pub const BYTE_POINTER_SIZE: Byte = Byte(8);

#[derive(Clone, Copy, PartialEq, PartialOrd, Eq, Ord, Debug)]
pub struct Byte(u64);

impl Byte {
    pub const fn new(value: u64) -> Byte {
        Byte(value)
    }

    pub fn increment(&mut self) {
        self.0 += 1;
    }

    pub const fn to_usize(self) -> usize {
        self.0 as usize
    }

    pub const fn to_u64(self) -> u64 {
        self.0
    }

    pub const fn multiply(lhs: Byte, rhs: Byte) -> Byte {
        Byte(lhs.0 * rhs.0)
    }
}

impl std::fmt::Display for Byte {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.0)
    }
}

impl From<Byte> for usize {
    fn from(b: Byte) -> usize {
        b.0 as usize
    }
}

impl From<usize> for Byte {
    fn from(u: usize) -> Byte {
        Byte(u as u64)
    }
}

impl From<Byte> for u64 {
    fn from(b: Byte) -> u64 {
        b.0
    }
}

impl From<u64> for Byte {
    fn from(u: u64) -> Byte {
        Byte(u)
    }
}

impl std::ops::Add for Byte {
    type Output = Self;

    fn add(self, other: Self) -> Self::Output {
        Self(self.0 + other.0)
    }
}

impl std::ops::AddAssign for Byte {
    fn add_assign(&mut self, other: Self) {
        self.0 += other.0
    }
}

impl std::ops::Sub for Byte {
    type Output = Self;

    fn sub(self, other: Self) -> Self::Output {
        Self(self.0 - other.0)
    }
}

impl std::ops::Mul for Byte {
    type Output = Self;

    fn mul(self, other: Self) -> Self::Output {
        Self(self.0 * other.0)
    }
}

impl std::ops::Div for Byte {
    type Output = Self;

    fn div(self, other: Self) -> Self::Output {
        Self(self.0 / other.0)
    }
}

impl std::ops::Rem for Byte {
    type Output = Self;

    fn rem(self, other: Self) -> Self::Output {
        Self(self.0 % other.0)
    }
}
