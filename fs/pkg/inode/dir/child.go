package dir

import (
	"fmt"
	"log"

	. "github.com/weberc2/mono/fs/pkg/types"
)

func CreateChild(
	fs *FileSystem,
	dir Ino,
	name []byte,
	fileType FileType,
	out *Inode,
) error {
	var dirInode Inode
	if err := fs.InodeStore.Get(dir, &dirInode); err != nil {
		return fmt.Errorf(
			"creating child `%s` in dir `%d`: %w",
			name,
			dir,
			err,
		)
	}

	if dirInode.FileType != FileTypeDir {
		return fmt.Errorf(
			"creating child `%s` in dir `%d`: %w",
			name,
			dir,
			NotADirErr,
		)
	}

	ino, ok := fs.InoAllocator.Alloc()
	if !ok {
		return fmt.Errorf(
			"creating child `%s` in dir `%d`: %w",
			name,
			dir,
			OutOfInosErr,
		)
	}

	log.Printf("creating child with ino %d", ino)
	if err := InitInode(
		fs,
		&dirInode,
		out,
		ino,
		fileType,
	); err != nil {
		return fmt.Errorf(
			"creating child inode named `%s` with type `%s` in dir `%d`: %w",
			name,
			fileType,
			dir,
			err,
		)
	}
	log.Printf("child has ino %d", out.Ino)

	if err := AddEntry(fs, &dirInode, out, name); err != nil {
		return fmt.Errorf(
			"creating child named `%s` with type `%s` in dir `%d`: %w",
			name,
			fileType,
			dir,
			err,
		)
	}

	return nil
}

const (
	OutOfInosErr ConstError = "out of inos"
)
