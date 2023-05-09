package dir

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/weberc2/mono/fs/pkg/encode"
	"github.com/weberc2/mono/fs/pkg/inode/data"
	. "github.com/weberc2/mono/fs/pkg/types"
)

func ReadNext(fs *FileSystem, handle *Handle, info *FileInfo) error {
	var inode Inode
	if err := fs.InodeStore.Get(handle.ino, &inode); err != nil {
		return fmt.Errorf(
			"reading entry from `%d` at offset `%d`: %w",
			handle.ino,
			handle.offset,
			err,
		)
	}

	r := fs.ReadWriter.Reader()
	var entry DirEntry

	// loop until we either run out of entries in the dir (if so, return EOF)
	for handle.offset < inode.Size {
		log.Printf("handle.offset: %d; inode.Size: %d", handle.offset, inode.Size)
		if err := ReadEntry(r, &inode, handle.offset, &entry); err != nil {
			return fmt.Errorf(
				"reading entry from `%d` at offset `%d`: %w",
				handle.ino,
				handle.offset,
				err,
			)
		}

		handle.offset += encode.DirEntrySize(&entry)

		// if the entry is nil, skip it
		if entry.Ino == InoNil {
			continue
		}

		// otherwise, populate the fields in `info`
		if entry.FileType != FileTypeInvalid {
			info.FileType = entry.FileType
		} else {
			// NB: I'm not sure if we would ever get here.
			log.Printf(
				"WARN unexpected dir entry with invalid file type for ino "+
					"`%d`; fetching file type from inode in inode store",
				entry.Ino,
			)
			var tmp Inode
			if err := fs.InodeStore.Get(entry.Ino, &tmp); err != nil {
				return fmt.Errorf(
					"reading entry from `%d` at offset `%d`: "+
						"entry for ino `%d` has invalid file type; "+
						"fetching file type from inode store: %w",
					handle.ino,
					handle.offset,
					entry.Ino,
					err,
				)
			}
			info.FileType = tmp.FileType
		}
		info.Ino = entry.Ino
		info.Name = entry.Name
		return nil
	}

	return io.EOF
}

type Handle struct {
	ino    Ino
	offset Byte
}

type FileInfo struct {
	Ino      Ino
	FileType FileType
	Name     []byte
}

func (fi *FileInfo) Equal(other *FileInfo) bool {
	return fi.Ino == other.Ino && fi.FileType == other.FileType &&
		bytes.Equal(fi.Name, other.Name)
}

func (fi *FileInfo) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Ino      Ino
		FileType string
		Name     string
	}{
		Ino:      fi.Ino,
		FileType: fi.FileType.String(),
		Name:     string(fi.Name),
	})
}

func ReadEntry(
	reader data.Reader,
	inode *Inode,
	offset Byte,
	out *DirEntry,
) error {
	buf := new([encode.DirEntryHeaderSize]byte)
	if _, err := reader.Read(
		inode,
		offset,
		buf[:],
	); err != nil {
		return fmt.Errorf(
			"reading direntry for inode `%d` at offset `%d`: %w",
			inode.Ino,
			offset,
			err,
		)
	}

	encode.DecodeDirEntryHeader(out, buf)

	out.Name = make([]byte, out.NameLen)
	if _, err := reader.Read(
		inode,
		offset+encode.DirEntryHeaderSize,
		out.Name,
	); err != nil {
		return fmt.Errorf(
			"reading direntry for inode `%d` at offset `%d`: %w",
			inode.Ino,
			offset,
			err,
		)
	}

	return nil
}
