package dir

import (
	"encoding/json"
	"fmt"
	stdio "io"
	"testing"

	"github.com/weberc2/mono/fs/pkg/alloc"
	"github.com/weberc2/mono/fs/pkg/inode/store"
	"github.com/weberc2/mono/fs/pkg/io"
	. "github.com/weberc2/mono/fs/pkg/types"
)

func TestAdd(t *testing.T) {
	type testCase struct {
		name          string
		state         FileSystem
		inputDir      Inode
		inputEntry    Inode
		inputName     string
		wantedDir     Inode
		wantedEntry   Inode
		wantedError   func(err error) error
		wantedEntries []FileInfo
	}

	testCases := []testCase{func() testCase {
		fs := defaultFileSystem()
		root := getRoot(&fs)
		entry := newInode(&fs, &root, FileTypeRegular)
		name := "entry"

		return testCase{
			name:       "simple",
			state:      fs,
			inputDir:   root,
			inputEntry: entry,
			inputName:  name,
			wantedEntries: []FileInfo{{
				Ino:      root.Ino,
				FileType: FileTypeDir,
				Name:     []byte("."),
			}, {
				Ino:      entry.Ino,
				FileType: entry.FileType,
				Name:     []byte(name),
			}},
		}
	}()}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if err := AddEntry(
				&tc.state,
				&tc.inputDir,
				&tc.inputEntry,
				[]byte(tc.name),
			); err != nil {
				if tc.wantedError == nil {
					t.Fatalf("AddEntry(): unexpected err: %v", err)
				} else {
					if err := tc.wantedError(err); err != nil {
						t.Fatal(err)
					}
				}
			}

			if tc.inputDir != tc.wantedDir {
				wanted, err := json.Marshal(tc.wantedDir)
				if err != nil {
					t.Fatalf("marshaling inode: %v", err)
				}
				found, err := json.Marshal(tc.inputDir)
				if err != nil {
					t.Fatalf("marshaling inode: %v", err)
				}
				t.Fatalf("wanted dir inode `%s`; found `%s`", wanted, found)
			}

			if tc.inputEntry != tc.wantedEntry {
				wanted, err := json.Marshal(tc.wantedEntry)
				if err != nil {
					t.Fatalf("marshaling inode: %v", err)
				}
				found, err := json.Marshal(tc.inputEntry)
				if err != nil {
					t.Fatalf("marshaling inode: %v", err)
				}
				t.Fatalf("wanted entry inode `%s`; found `%s`", wanted, found)
			}

			var h Handle
			if err := Open(&tc.state, tc.inputDir.Ino, &h); err != nil {
				t.Fatalf("Open(): unexpected err: %v", err)
			}

			var found []FileInfo
			for {
				var info FileInfo
				if err := ReadNext(&tc.state, &h, &info); err != nil {
					if err == stdio.EOF {
						break
					}
					t.Fatalf("ReadNext(): unexpected err: %v", err)
				}
				found = append(found, info)
			}

			if len(found) != len(tc.wantedEntries) {
				wanted, err := json.Marshal(tc.wantedEntries)
				if err != nil {
					t.Fatalf("marshaling []FileInfo: %v", err)
				}
				found, err := json.Marshal(found)
				if err != nil {
					t.Fatalf("marshaling []FileInfo: %v", err)
				}
				t.Fatalf("wanted []FileInfo `%s`; found `%s`", wanted, found)
			}

			for i := range tc.wantedEntries {
				if !tc.wantedEntries[i].Equal(&found[i]) {
					wanted, err := json.Marshal(tc.wantedEntries[i])
					if err != nil {
						t.Fatalf("marshaling FileInfo: %v", err)
					}
					found, err := json.Marshal(found)
					if err != nil {
						t.Fatalf("marshaling FileInfo: %v", err)
					}
					t.Fatalf("wanted FileInfo `%s`; found `%s`", wanted, found)
				}
			}
		})
	}
}

func defaultFileSystem() FileSystem {
	volume := io.NewBuffer(make([]byte, 1024*BlockSize))
	blockVolume := io.NewOffsetVolume(volume, 10*BlockSize)
	blockAllocator := alloc.BlockAllocator{alloc.New(BlockSize)}
	inoAllocator := alloc.InoAllocator{alloc.New(BlockSize)}
	inodeStore := store.NewVolumeInodeStore(volume)
	var fs FileSystem
	fs.Init(blockAllocator, inoAllocator, blockVolume, inodeStore)
	var root Inode
	if err := InitRoot(&fs, &root); err != nil {
		panic(fmt.Sprintf("initializing root: %v", err))
	}
	return fs
}

func newInode(fs *FileSystem, parent *Inode, fileType FileType) Inode {
	var inode Inode
	ino, ok := fs.InoAllocator.Alloc()
	if !ok {
		panic("out of inodes")
	}
	if err := InitInode(fs, parent, &inode, ino, fileType); err != nil {
		panic(fmt.Sprintf("initializing inode: %v", err))
	}
	return inode
}

func getRoot(fs *FileSystem) Inode {
	var root Inode
	if err := fs.InodeStore.Get(InoRoot, &root); err != nil {
		panic(fmt.Sprintf("getting root inode: %v", err))
	}
	return root
}
