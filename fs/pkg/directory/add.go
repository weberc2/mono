package directory

import (
	"fmt"
	"log"

	"github.com/weberc2/mono/fs/pkg/encode"
	. "github.com/weberc2/mono/fs/pkg/types"
)

func AddEntry(
	fs *FileSystem,
	dir *Inode,
	entry *Inode,
	name string,
) error {
	if dir.Ino == 3 {
		log.Printf(
			"writing direntry `%s` to inode `%d`",
			name,
			dir.Ino,
		)
	}
	var offset, lastOffset Byte
	reader := fs.ReadWriter.Reader()
	var dirEntry DirEntry
	for offset < dir.Size {
		if err := ReadEntry(
			reader,
			dir,
			offset,
			&dirEntry,
		); err != nil {
			return fmt.Errorf(
				"adding entry `%d` to dir `%d` with name `%s`: %w",
				entry.Ino,
				dir.Ino,
				name,
				err,
			)
		}

		// if the new entry can fit in the previous entry's free space, then
		// insert it there.
		newEntrySize := encode.DirEntrySize(uint8(len(name)))
		if encode.DirEntryFreeSpace(&dirEntry) >= newEntrySize {
			// insert the new entry at `offset + dirEntry.FreeSpace()` and
			// update the previous entry's reclen field. This order is important
			// so if an error happens we don't shrink the record with
			// potentially garbage data in the next record (potentially making
			// subsequent entries unreachable).
			if err := InsertEntry(
				fs,
				dir,
				entry,
				name,
				offset+encode.DirEntrySize(dirEntry.NameLen),
			); err != nil {
				return fmt.Errorf(
					"adding inode `%d` as entry in inode `%d` named `%s`: %w",
					entry.Ino,
					dir.Ino,
					name,
					err,
				)
			}

			var buf [2]byte
			encode.EncodeDirEntryRecLen(
				uint16(encode.DirEntrySize(dirEntry.NameLen)),
				&buf,
			)
			if _, err := fs.ReadWriter.Write(
				dir,
				offset+encode.DirEntryRecLenStart,
				buf[:],
			); err != nil {
				return fmt.Errorf(
					"adding inode `%d` as entry in inode `%d` named `%s`: "+
						"updating previous entry's record length: %w",
					entry.Ino,
					dir.Ino,
					name,
					err,
				)
			}

			return nil
		}

		lastOffset = offset
		offset += Byte(dirEntry.RecLen)
		if dirEntry.Ino == InoNil {
			break
		}
	}

	if err := InsertEntry(
		fs,
		dir,
		entry,
		name,
		lastOffset+encode.DirEntrySize(dirEntry.NameLen),
	); err != nil {
		return fmt.Errorf(
			"adding inode `%d` as entry in inode `%d` named `%s`: %w",
			entry.Ino,
			dir.Ino,
			name,
			err,
		)
	}

	// if there was some spare capacity, then we will have written into it, and
	// thus we need to update the last direntry to shrink its reclen
	if encode.DirEntryFreeSpace(&dirEntry) > 0 {
		var buf [2]byte
		encode.EncodeDirEntryRecLen(
			uint16(encode.DirEntrySize(dirEntry.NameLen)),
			&buf,
		)
		if _, err := fs.ReadWriter.Write(
			dir,
			lastOffset+encode.DirEntryRecLenStart,
			buf[:],
		); err != nil {
			return fmt.Errorf(
				"adding inode `%d` as entry in inode `%d` named `%s`: "+
					"updating previous entry's record length: %w",
				entry.Ino,
				dir.Ino,
				name,
				err,
			)
		}
	}
	return nil
}

func align4(x Byte) Byte {
	return (x + 0b11) &^ 0b11
}
