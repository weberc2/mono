mod manager;
mod physicalreader;
mod physicalwriter;

#[cfg(test)]
mod physicalwriter_test;

#[cfg(test)]
mod physicalreader_test;

#[cfg(test)]
mod manager_test;

pub use manager::{Manager, Reader, Writer};
pub use physicalreader::PhysicalReader;
pub use physicalwriter::PhysicalWriter;
use thiserror::Error;

#[derive(Error, Debug)]
pub enum IndirectionError {
    #[error("io error")]
    Io(#[from] std::io::Error),
    #[error("out of free blocks")]
    OutOfBlocks,
    #[error("block exceeds the maximum number of blocks in an inode")]
    OutOfRange,
    #[error("{0}: {1}")]
    Annotated(String, Box<IndirectionError>),
}

pub type Result<T> = std::result::Result<T, IndirectionError>;

#[macro_export]
macro_rules! annotate {
    ($r:expr, $f:expr, $($arg:tt)*) => {{
        let r: std::result::Result<_, _> = $r;
        r.map_err(|e| IndirectionError::Annotated(format!($f, $($arg)*), Box::new(IndirectionError::from(e))))
    }};
}
