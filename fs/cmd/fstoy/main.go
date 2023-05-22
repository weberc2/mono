package main

import (
	"encoding/json"
	"fmt"
	stdio "io"
	"log"

	"github.com/weberc2/mono/fs/pkg/alloc"
	"github.com/weberc2/mono/fs/pkg/directory"
	"github.com/weberc2/mono/fs/pkg/filesystem"
	"github.com/weberc2/mono/fs/pkg/inode/store"
	"github.com/weberc2/mono/fs/pkg/io"
	. "github.com/weberc2/mono/fs/pkg/types"
)

func main() {
	volume := io.NewBuffer(make([]byte, 1024*1024))
	var fs directory.FileSystem
	fs.Init(
		alloc.BlockAllocator{Allocator: alloc.New(1024)},
		alloc.InoAllocator{Allocator: alloc.New(1024)},
		io.NewOffsetVolume(volume, 1024*10),
		newInodeStore(volume),
	)

	var root Inode
	log.Printf("before: %s", jsonify(&root))

	if err := directory.InitRootDirectory(&fs, &root); err != nil {
		log.Fatalf("initializing root inode: %v", err)
	}

	if err := fs.InodeStore.Get(InoRoot, &root); err != nil {
		log.Fatalf("fetching root inode: %v", err)
	}
	log.Printf("after: %s", jsonify(&root))

	log.Printf("root.Size: %d", root.Size)
	var helloInode, worldInode Inode
	if err := directory.CreateChild(
		&fs,
		root.Ino,
		"hello",
		FileTypeRegular,
		&helloInode,
	); err != nil {
		log.Fatalf("creating child inode: %v", err)
	}

	if err := directory.CreateChild(
		&fs,
		root.Ino,
		"world",
		FileTypeDir,
		&worldInode,
	); err != nil {
		log.Fatalf("creating child inode: %v", err)
	}

	log.Printf("after: %s", jsonify(&root))

	var fooInode Inode
	if err := directory.CreateChild(
		&fs,
		worldInode.Ino,
		"foo",
		FileTypeRegular,
		&fooInode,
	); err != nil {
		log.Fatalf("creating child inode: %v", err)
	}

	log.Printf("world: %s", jsonify(&worldInode))

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

	if err := listFiles(&fs, InoRoot); err != nil {
		log.Fatal(err)
	}

	if err := listFiles(&fs, worldInode.Ino); err != nil {
		log.Fatal(err)
	}

	var info directory.FileInfo
	if err := directory.Lookup(
		&fs,
		InoRoot,
		"world",
		&info,
	); err != nil {
		log.Fatal(err)
	}
	log.Printf("result: %s", jsonify(&info))

	if err := directory.Lookup(
		&fs,
		info.Ino,
		"foo",
		&info,
	); err != nil {
		log.Fatal(err)
	}
	log.Printf("result: %s", jsonify(&info))

	if err := filesystem.Lookup(&fs, "/world", &info); err != nil {
		log.Fatal(err)
	}
	log.Printf("result: %s", jsonify(&info))
}

func listFiles(fs *directory.FileSystem, ino Ino) error {
	var h directory.Handle
	if err := directory.Open(fs, ino, &h); err != nil {
		return err
	}

	for {
		var info directory.FileInfo
		if err := directory.ReadNext(fs, &h, &info); err != nil {
			if err == stdio.EOF {
				return nil
			}
			return err
		}
		log.Printf("info: %s", jsonify(&info))
	}
}

func newInodeStore(volume io.Volume) *store.CachingInodeStore {
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
