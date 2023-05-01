package fs

func IsFastSymlink(fs *FileSystem, inode *Inode) bool {
	if inode.Mode.Type != FileTypeSymlink {
		return false
	}
	if inode.ACL != 0 {
		return inode.Size == fs.Superblock.BlockSize
	}
	return inode.Size == 0
}
