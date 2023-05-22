package directory

import (
	"encoding/json"
	"fmt"
	stdio "io"
	"log"
	"testing"

	"github.com/weberc2/mono/fs/pkg/alloc"
	"github.com/weberc2/mono/fs/pkg/encode"
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
		hook          func()
	}

	testCases := []testCase{func() testCase {
		fs := defaultFileSystem()
		root := getRoot(&fs)
		entry := newInode(&fs, &root, FileTypeRegular)
		name := "entry"

		wantedRoot := root
		wantedRoot.Size += encode.DirEntrySize(uint8(len(name)))

		wantedEntry := entry
		wantedEntry.LinksCount++

		return testCase{
			name:        "append",
			state:       fs,
			inputDir:    root,
			inputEntry:  entry,
			inputName:   name,
			wantedDir:   wantedRoot,
			wantedEntry: wantedEntry,
			wantedEntries: []FileInfo{{
				Ino:      root.Ino,
				FileType: FileTypeDir,
				Name:     ".",
			}, {
				Ino:      entry.Ino,
				FileType: entry.FileType,
				Name:     name,
			}},
			hook: hook,
		}
	}(), func() testCase {
		fs := defaultFileSystem()
		root := getRoot(&fs)

		// Given the root has two DirEntries: the first being the '.' entry and
		// the second being an entry with sufficient free space to fit the new
		// entry.
		bigEntry := newInode(&fs, &root, FileTypeRegular)
		bigEntryName := "bigentry"
		const bigEntryRecLen = 512 // plenty of space for new entry
		if err := WriteEntry(
			fs.ReadWriter.Writer(),
			&root,
			encode.DirEntrySize(1), // right after the `.` entry
			&DirEntry{
				Ino:      bigEntry.Ino,
				NameLen:  uint8(len(bigEntryName)),
				FileType: bigEntry.FileType,
				RecLen:   bigEntryRecLen,
				Name:     bigEntryName,
			},
		); err != nil {
			t.Fatalf("writing bigentry: unexpected err: %v", err)
		}
		if err := fs.InodeStore.Put(&root); err != nil {
			t.Fatalf("storing updated root inode: unexpected err: %v", err)
		}
		newEntry := newInode(&fs, &root, FileTypeRegular)
		name := "entry"

		wantedRoot := root
		wantedRoot.Size += encode.DirEntrySize(uint8(len(name)))

		wantedEntry := newEntry
		wantedEntry.LinksCount++

		return testCase{
			name:        "insert",
			state:       fs,
			inputDir:    root,
			inputEntry:  newEntry,
			inputName:   name,
			wantedDir:   wantedRoot,
			wantedEntry: wantedEntry,
			wantedEntries: []FileInfo{{
				Ino:      root.Ino,
				FileType: FileTypeDir,
				Name:     ".",
			}, {
				Ino:      bigEntry.Ino,
				FileType: bigEntry.FileType,
				Name:     bigEntryName,
			}, {
				Ino:      newEntry.Ino,
				FileType: newEntry.FileType,
				Name:     name,
			}},
		}
	}(), func() testCase {
		fs := defaultFileSystem()
		root := getRoot(&fs)

		// Given the root has two DirEntries: the first being the '.' entry and
		// the second being an entry with sufficient free space to fit the new
		// entry.
		smallEntry := newInode(&fs, &root, FileTypeRegular)
		smallEntryName := "smallentry"
		// only leave a free space of a few bytes
		smallEntryRecLen := uint16(
			encode.DirEntrySize(uint8(len(smallEntryName))) + 5,
		)
		if err := WriteEntry(
			fs.ReadWriter.Writer(),
			&root,
			encode.DirEntrySize(1), // right after the `.` entry
			&DirEntry{
				Ino:      smallEntry.Ino,
				NameLen:  uint8(len(smallEntryName)),
				FileType: smallEntry.FileType,
				RecLen:   smallEntryRecLen,
				Name:     smallEntryName,
			},
		); err != nil {
			t.Fatalf("writing bigentry: unexpected err: %v", err)
		}
		if err := fs.InodeStore.Put(&root); err != nil {
			t.Fatalf("storing updated root inode: unexpected err: %v", err)
		}
		newEntry := newInode(&fs, &root, FileTypeRegular)
		name := "entry"

		wantedRoot := root
		wantedRoot.Size += encode.DirEntrySize(uint8(len(name)))

		wantedEntry := newEntry
		wantedEntry.LinksCount++

		return testCase{
			name:        "try-insert",
			state:       fs,
			inputDir:    root,
			inputEntry:  newEntry,
			inputName:   name,
			wantedDir:   wantedRoot,
			wantedEntry: wantedEntry,
			wantedEntries: []FileInfo{{
				Ino:      root.Ino,
				FileType: FileTypeDir,
				Name:     ".",
			}, {
				Ino:      smallEntry.Ino,
				FileType: smallEntry.FileType,
				Name:     smallEntryName,
			}, {
				Ino:      newEntry.Ino,
				FileType: newEntry.FileType,
				Name:     name,
			}},
		}
	}()}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.hook != nil {
				tc.hook()
			}
			if err := AddEntry(
				&tc.state,
				&tc.inputDir,
				&tc.inputEntry,
				tc.inputName,
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
	if err := InitRootDirectory(&fs, &root); err != nil {
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

func hook() {
	log.Println("hello")
}
