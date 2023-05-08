package dir

import (
	"bytes"
	"fmt"
	"log"

	"github.com/weberc2/mono/fs/pkg/encode"
	"github.com/weberc2/mono/fs/pkg/math"
	. "github.com/weberc2/mono/fs/pkg/types"
)

func AddEntry(
	fs *FileSystem,
	dir *Inode,
	entry *Inode,
	name []byte,
) error {
	reader := fs.ReadWriter.Reader()
	writer := fs.ReadWriter.Writer()

	if dir.FileType != FileTypeDir {
		return fmt.Errorf(
			"adding inode `%d` as entry in inode `%d` named `%s`: %w",
			entry.Ino,
			dir.Ino,
			name,
			NotADirErr,
		)
	}

	entrySize := encode.DirEntryHeaderSize + Byte(len(name))

	var placeForEntry *FreeSpace
	var offset Byte
	var lastOffset Byte
	for offset < dir.Size {
		log.Printf("offset=%d; dir.Size=%d", offset, dir.Size)
		var dirEntry DirEntry
		if err := ReadEntry(&reader, dir, offset, &dirEntry); err != nil {
			return fmt.Errorf(
				"adding inode `%d` as entry in inode `%d` named `%s`: %w",
				entry.Ino,
				dir.Ino,
				name,
				err,
			)
		}

		nextOffset := offset + encode.DirEntryHeaderSize +
			Byte(dirEntry.NameLen)
		if bytes.Equal(dirEntry.Name, name) {
			if dirEntry.Ino == entry.Ino {
				return nil
			}

			if err := WriteEntry(
				&writer,
				dir,
				offset,
				&DirEntry{
					Ino:      dirEntry.Ino,
					FileType: entry.FileType,
					NameLen:  dirEntry.NameLen,
					Name:     dirEntry.Name,
				},
			); err != nil {
				return fmt.Errorf(
					"adding inode `%d` as entry in inode `%d` named `%s`: %w",
					entry.Ino,
					dir.Ino,
					name,
					err,
				)
			}

			entry.LinksCount++
			var oldInode Inode
			if err := fs.InodeStore.Get(dirEntry.Ino, &oldInode); err != nil {
				return fmt.Errorf(
					"adding inode `%d` as entry in inode `%d` named `%s`: %w",
					entry.Ino,
					dir.Ino,
					name,
					err,
				)
			}

			if _, err := UnlinkInode(&reader, &oldInode); err != nil {
				return fmt.Errorf(
					"adding inode `%d` as entry in inode `%d` named `%s`: %w",
					entry.Ino,
					dir.Ino,
					name,
					err,
				)
			}

			return nil
		}

		freeOffset := align4(
			offset + encode.DirEntryHeaderSize + Byte(dirEntry.NameLen),
		)
		spaceInBlock := (offset/BlockSize*BlockSize + BlockSize) - offset
		freeSize := math.Min(nextOffset-freeOffset, spaceInBlock)

		if placeForEntry == nil && freeSize >= entrySize {
			placeForEntry = &FreeSpace{
				Offset:     freeOffset,
				PrevOffset: offset,
				NextOffset: nextOffset,
			}
		}

		lastOffset = offset
		offset = nextOffset
	}

	if err := InsertEntry(
		&writer,
		fs.InodeStore,
		dir,
		entry,
		name,
		placeForEntry,
		lastOffset,
	); err != nil {
		return fmt.Errorf(
			"adding inode `%d` as entry in inode `%d` named `%s`: %w",
			entry.Ino,
			dir.Ino,
			name,
			err,
		)
	}
	return nil
}

func align4(x Byte) Byte {
	return (x + 0b11) &^ 0b11
}
