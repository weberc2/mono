use indirectmanager::Wrapper;
use inodeblock::PhysicalReader;

mod bitmap;
mod block;
mod byte;
mod cachinginodestore;
mod encode;
mod indirectmanager;
mod inode;
mod inodeblock;
mod inodedata;
mod inodestore;
mod io;
mod volumeinodestore;
mod write;

fn main() {
    let mut volume = vec![0 as u8; 1024 * 1024];
    let mut indirect_reader = Wrapper::new(&mut volume);
    let physical_reader = PhysicalReader::new(indirect_reader)
}
