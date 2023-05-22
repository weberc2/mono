package file

import (
	"fmt"

	"github.com/weberc2/mono/fs/pkg/directory"
	. "github.com/weberc2/mono/fs/pkg/types"
)

type FileSystem = directory.FileSystem

type Handle struct {
	ino Ino
}

func Open(fs *FileSystem, ino Ino) (Handle, error) {
	var inode Inode
	if err := fs.InodeStore.Get(ino, &inode); err != nil {
		return Handle{}, fmt.Errorf(
			"opening inode `%d` as regular file: %w",
			ino,
			err,
		)
	}

	if inode.FileType != FileTypeRegular {
		return Handle{}, fmt.Errorf(
			"opening inode `%d` as regular file: %w",
			ino,
			NotARegularFileErr,
		)
	}

	return Handle{ino}, nil
}

func Read(fs *FileSystem, h Handle, offset Byte, b []byte) (Byte, error) {
	var inode Inode
	if err := fs.InodeStore.Get(h.ino, &inode); err != nil {
		return 0, fmt.Errorf("reading data from file `%d`: %w", h.ino, err)
	}

	n, err := fs.ReadWriter.Read(&inode, offset, b)
	if err != nil {
		return n, fmt.Errorf("reading data from file `%d`: %w", h.ino, err)
	}

	return n, nil
}

func Write(fs *FileSystem, h Handle, offset Byte, b []byte) (Byte, error) {
	var inode Inode
	if err := fs.InodeStore.Get(h.ino, &inode); err != nil {
		return 0, fmt.Errorf("writing data from file `%d`: %w", h.ino, err)
	}

	n, err := fs.ReadWriter.Write(&inode, offset, b)
	if err != nil {
		return n, fmt.Errorf("writing data from file `%d`: %w", h.ino, err)
	}

	return n, nil
}

func Close(fs *FileSystem, h Handle) error {
	return fs.InodeStore.Flush(h.ino)
}

const (
	NotARegularFileErr ConstError = "not a regular file"
)
