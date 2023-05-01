package fs

import (
	"fmt"
)

func ReadSuperblock(fs *FileSystem) error {
	var buf [SuperblockSize]byte
	if err := ReadAt(fs.Volume, SuperblockOffset, buf[:]); err != nil {
		return fmt.Errorf("reading superblock: %w", err)
	}
	if err := DecodeSuperblock(&fs.Superblock, &buf); err != nil {
		return fmt.Errorf("reading superblock: %w", err)
	}
	return nil
}

func ReadDescriptor(fs *FileSystem) error {
	var buf [DescriptorUsedDirsCountSize]byte
	if err := ReadAt(
		fs.Volume,
		fs.Superblock.DescriptorOffset(),
		buf[:],
	); err != nil {
		return fmt.Errorf(
			"reading descriptor: reading count of used dirs: %w",
			err,
		)
	}
	fs.Descriptor.UsedDirsCount = getU64(buf[:])

	if err := ReadAt(
		fs.Volume,
		fs.Superblock.BlockBitmapOffset(),
		fs.Descriptor.BlockBitmap,
	); err != nil {
		return fmt.Errorf("reading descriptor: reading block bitmap: %w", err)
	}

	if err := ReadAt(
		fs.Volume,
		fs.Superblock.InodeBitmapOffset(),
		fs.Descriptor.InodeBitmap,
	); err != nil {
		return fmt.Errorf("reading descriptor: reading inode bitmap: %w", err)
	}

	return nil
}

func ReadInode(fs *FileSystem, ino Ino, out *Inode) error {
	var buf [InodeSize]byte
	if err := ReadAt(
		fs.Volume,
		fs.Superblock.InodeOffset(ino),
		buf[:],
	); err != nil {
		return fmt.Errorf("reading inode `%d`: %w", ino, err)
	}
	if err := DecodeInode(out, &buf); err != nil {
		return fmt.Errorf("reading inode `%d`: %w", ino, err)
	}
	out.Ino = ino
	return nil
}
