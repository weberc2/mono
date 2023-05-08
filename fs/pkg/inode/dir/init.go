package dir

import (
	"fmt"

	"github.com/weberc2/mono/fs/pkg/encode"
	. "github.com/weberc2/mono/fs/pkg/types"
)

func InitRoot(fs *FileSystem, root *Inode) error {
	b := new([encode.DirEntryHeaderSize + Byte(len("."))]byte)

	root.FileType = FileTypeDir
	root.Ino = InoRoot
	fs.InoAllocator.Reserve(InoRoot)

	dotEntry := DirEntry{
		Ino:      root.Ino,
		FileType: FileTypeDir,
		NameLen:  uint8(len(".")),
		Name:     []byte("."),
	}

	encode.EncodeDirEntryHeader(
		&dotEntry,
		(*[encode.DirEntryHeaderSize]byte)(b[:]),
	)
	b[encode.DirEntryHeaderSize] = '.'

	if _, err := fs.ReadWriter.Write(root, 0, b[:]); err != nil {
		return fmt.Errorf("initializing root dir `%d`: %w", root.Ino, err)
	}
	root.LinksCount++
	if err := fs.InodeStore.Put(root); err != nil {
		return fmt.Errorf(
			"initializing root dir `%d`: storing dir inode: %w",
			root.Ino,
			err,
		)
	}

	// NB: Potentially update a superblock or other metadata and mark it dirty
	// (for example, ext2 updates a counter of used directories on a group
	// descriptor and marks the group descriptor dirty).
	return nil
}

func InitInode(
	fs *FileSystem,
	parent *Inode,
	entry *Inode,
	ino Ino,
	fileType FileType,
) error {
	*entry = Inode{
		Ino:      ino,
		FileType: fileType,
	}

	if fileType == FileTypeDir {
		if err := InitDir(fs, parent, entry); err != nil {
			return fmt.Errorf(
				"initializing new directory inode `%d` in parent `%d`: %w",
				ino,
				parent.Ino,
				err,
			)
		}
	} else {
		// Only need to do this if it's not a dir--otherwise InitDir() will
		// update the inode store.
		if err := fs.InodeStore.Put(entry); err != nil {
			return fmt.Errorf(
				"initializing new %s inode `%d` in parent `%d`: %w",
				fileType,
				ino,
				parent.Ino,
				err,
			)
		}
	}

	return nil
}

func InitDir(fs *FileSystem, parent *Inode, dir *Inode) error {
	dotDotOffset := align4(encode.DirEntryHeaderSize + Byte(len("..")))
	b := new([BlockSize]byte)

	dotEntry := DirEntry{
		Ino:      dir.Ino,
		FileType: FileTypeDir,
		NameLen:  uint8(len(".")),
		Name:     []byte("."),
	}

	dotDotEntry := DirEntry{
		Ino:      parent.Ino,
		FileType: FileTypeDir,
		NameLen:  uint8(len("..")),
		Name:     []byte(".."),
	}

	encode.EncodeDirEntryHeader(
		&dotEntry,
		(*[encode.DirEntryHeaderSize]byte)(b[:]),
	)
	b[encode.DirEntryHeaderSize] = '.'

	encode.EncodeDirEntryHeader(
		&dotDotEntry,
		(*[encode.DirEntryHeaderSize]byte)(b[dotDotOffset:]),
	)
	copy(b[dotDotOffset+encode.DirEntryHeaderSize:], "..")

	if _, err := fs.ReadWriter.Write(dir, 0, b[:]); err != nil {
		return fmt.Errorf(
			"initializing dir `%d` with parent `%d`: %w",
			dir.Ino,
			parent.Ino,
			err,
		)
	}
	parent.LinksCount++
	dir.LinksCount++
	if err := fs.InodeStore.Put(parent); err != nil {
		return fmt.Errorf(
			"initializing dir `%d` with parent `%d`: storing parent inode: %w",
			dir.Ino,
			parent.Ino,
			err,
		)
	}
	if err := fs.InodeStore.Put(dir); err != nil {
		return fmt.Errorf(
			"initializing dir `%d` with parent `%d`: storing dir inode: %w",
			dir.Ino,
			parent.Ino,
			err,
		)
	}

	// NB: Potentially update a superblock or other metadata and mark it dirty
	// (for example, ext2 updates a counter of used directories on a group
	// descriptor and marks the group descriptor dirty).
	return nil
}
