package types

type InodeStore interface {
	Put(inode *Inode) error
	Get(ino Ino, output *Inode) error
}
