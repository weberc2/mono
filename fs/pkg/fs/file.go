package fs

import "fmt"

const (
	ErrNotRegular constErr = "not a regular file"
)

type FileHandle Ino

func OpenFile(fs *FileSystem, ino Ino) (FileHandle, error) {
	inode, err := GetInode(fs, ino)
	if err != nil {
		return 0, fmt.Errorf("opening file for ino `%d`: %w", ino, err)
	}
	if inode.Mode.Type == FileTypeRegular {
		return FileHandle(inode.Ino), nil
	}
	return 0, fmt.Errorf("opening file for ino `%d`: %w", ino, ErrNotRegular)
}
