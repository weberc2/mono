package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/weberc2/mono/fs/pkg/fs"
)

func main() {
	// if false {
	// f, err := os.Create("/tmp/test.img")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer f.Close()

	// if err := f.Truncate(1024 * 1024); err != nil {
	// 	log.Fatalf("truncating /tmp/test.img: %v", err)
	// }

	f := fs.NewBuffer(make([]byte, 1024*1024))
	if err := initFS(f); err != nil {
		panic(err)
	}
	if err := loadFS(f); err != nil {
		panic(err)
	}

	// } else {
	// 	f, err := os.Open("/tmp/test.img")
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	defer f.Close()

	//}
}

func loadFS(v io.ReadWriteSeeker) error {
	fileSystem := fs.FileSystem{
		Volume:     v,
		InodeCache: fs.NewCache(10),
		DirtyInos:  fs.NewInoSet(),
	}
	if err := fs.ReadSuperblock(&fileSystem); err != nil {
		return fmt.Errorf("loading fs: %w", err)
	}
	fileSystem.Descriptor = fs.NewDescriptor(
		fileSystem.Superblock.BlockCount,
		fileSystem.Superblock.InodeCount,
	)
	if err := fs.ReadDescriptor(&fileSystem); err != nil {
		log.Fatal(err)
		return fmt.Errorf("loading fs: %w", err)
	}

	fmt.Printf("superblock: %s\n", mustJSON(&fileSystem.Superblock))
	fmt.Printf("descriptor: %s\n", fileSystem.Descriptor.Debug())

	root, err := fs.GetInode(&fileSystem, fs.InoRoot)
	if err != nil {
		return fmt.Errorf("loading fs: fetching root inode: %w", err)
	}

	fmt.Printf("%s\n", mustJSON(root))

	if err := fs.MakeDir(&fileSystem, &root, &root); err != nil {
		return fmt.Errorf("initializing root inode: %w", err)
	}

	if _, err := fs.MakeInodeInDir(
		&fileSystem,
		&fs.InodeInDirParams{
			DirIno: fs.InoRoot,
			Name:   "hello",
			Mode: fs.Mode{
				Type:         fs.FileTypeRegular,
				SUID:         false,
				SGID:         false,
				Sticky:       false,
				AccessRights: 0644,
			},
		},
	); err != nil {
		return fmt.Errorf("making file `/hello`: %w", err)
	}

	return nil
}

func initFS(v io.ReadWriteSeeker) error {
	return fs.InitFileSystem(&fs.FileSystemParams{
		Volume:        v,
		BlockSize:     fs.DefaultBlockSize,
		Blocks:        1024,
		Inodes:        100,
		CacheCapacity: 10,
	})
}

func mustJSON(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		log.Fatalf("marshaling %#v to json: %v", v, err)
	}
	return data
}
