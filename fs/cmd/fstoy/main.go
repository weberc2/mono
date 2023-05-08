package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/weberc2/mono/fs/pkg/alloc"
	"github.com/weberc2/mono/fs/pkg/inode/dir"
	"github.com/weberc2/mono/fs/pkg/inode/store"
	"github.com/weberc2/mono/fs/pkg/io"
	. "github.com/weberc2/mono/fs/pkg/types"
)

func main() {
	volume := io.NewBuffer(make([]byte, 1024*1024))
	var fs dir.FileSystem
	fs.Init(
		alloc.BlockAllocator{Allocator: alloc.New(1024)},
		alloc.InoAllocator{Allocator: alloc.New(1024)},
		io.NewOffsetVolume(volume, 1024*10),
		newInodeStore(volume),
	)

	var root Inode
	log.Printf("before: %s", jsonify(&root))

	if err := dir.InitRoot(&fs, &root); err != nil {
		log.Fatalf("initializing root inode: %v", err)
	}

	if err := fs.InodeStore.Get(InoRoot, &root); err != nil {
		log.Fatalf("fetching root inode: %v", err)
	}
	log.Printf("after: %s", jsonify(&root))

	log.Printf("root.Size: %d", root.Size)
	var helloInode Inode
	if err := dir.CreateChild(
		&fs,
		root.Ino,
		[]byte("hello"),
		FileTypeRegular,
		&helloInode,
	); err != nil {
		log.Fatalf("creating child inode: %v", err)
	}

	log.Printf("after: %s", jsonify(&root))

	if _, err := fs.ReadWriter.Write(
		&helloInode,
		helloInode.Size,
		[]byte("hello, world"),
	); err != nil {
		log.Fatalf("writing: %v", err)
	}

	buf := make([]byte, 1024)
	if _, err := fs.ReadWriter.Read(&helloInode, 0, buf); err != nil {
		log.Fatalf("reading data: %v", err)
	}

	log.Printf("data: %s", buf)
}

func newInodeStore(volume io.Volume) InodeStore {
	const inodeTableOffset = 0
	inodeStoreOffsetVolume := io.NewOffsetVolume(volume, inodeTableOffset)
	inodeStoreBackend := store.NewVolumeInodeStore(inodeStoreOffsetVolume)
	return store.NewCachingInodeStore(inodeStoreBackend, 10)
}

func jsonify(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal data: %v", err))
	}
	return data
}
