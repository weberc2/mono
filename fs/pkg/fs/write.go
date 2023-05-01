package fs

import (
	"fmt"
)

func WriteSuperblock(fs *FileSystem) error {
	var buf [SuperblockSize]byte
	EncodeSuperblock(&fs.Superblock, &buf)
	if err := WriteAt(fs.Volume, SuperblockOffset, buf[:]); err != nil {
		return fmt.Errorf("writing superblock: %w", err)
	}
	return nil
}

func WriteDescriptor(fs *FileSystem) error {
	var buf [DescriptorUsedDirsCountSize]byte
	putU64(buf[:], fs.Descriptor.UsedDirsCount)
	if err := WriteAt(
		fs.Volume,
		fs.Superblock.DescriptorOffset(),
		buf[:],
	); err != nil {
		return fmt.Errorf(
			"writing descriptor: writing count of used dirs: %w",
			err,
		)
	}

	if err := WriteAt(
		fs.Volume,
		fs.Superblock.BlockBitmapOffset(),
		fs.Descriptor.BlockBitmap,
	); err != nil {
		return fmt.Errorf("writing descriptor: writing block bitmap: %w", err)
	}

	if err := WriteAt(
		fs.Volume,
		fs.Superblock.InodeBitmapOffset(),
		fs.Descriptor.InodeBitmap,
	); err != nil {
		return fmt.Errorf("writing descriptor: writing inode bitmap: %w", err)
	}

	return nil
}

func WriteInode(fs *FileSystem, inode *Inode) error {
	var buf [InodeSize]byte
	EncodeInode(inode, &buf)
	if err := WriteAt(
		fs.Volume,
		fs.Superblock.InodeOffset(inode.Ino),
		buf[:],
	); err != nil {
		return fmt.Errorf("writing inode `%d`: %w", inode.Ino, err)
	}
	delete(fs.DirtyInos, inode.Ino)
	return nil
}

func WriteDirEntry(fs *FileSystem, dir *Inode, entry *DirEntry) error {
	buf := make([]byte, DirEntrySize(Byte(len(entry.Name))))
	EncodeDirEntryHeader(&entry.Header, (*[SizeDirEntryHeader]byte)(buf))
	copy(buf[DirEntrySize(0):], entry.Name)
	if _, err := WriteInodeData(fs, dir, entry.NextOffset, buf); err != nil {
		return fmt.Errorf("writing dir entry for inode `%d`: %w", dir.Ino, err)
	}
	return nil
}
