package fs

import (
	"fmt"
	"log"
)

const (
	NotDirErr                constErr = "not a directory"
	EntryTooShortErr         constErr = "entry too short"
	FileNameTooLongErr       constErr = "file name too long"
	DirIsOwnParentErr        constErr = "directory's '..' entry points to itself"
	DirMissingParentEntryErr constErr = "directory has no '..' entry"
	DirMissingSelfEntryErr   constErr = "directory has no '.' entry"
)

type InodeInDirParams struct {
	DirIno Ino
	Name   string
	Mode   Mode
	Attr   FileAttr
}

func MakeInodeInDir(fs *FileSystem, params *InodeInDirParams) (Inode, error) {
	dirInode, err := GetInode(fs, params.DirIno)
	if err != nil {
		return Inode{}, fmt.Errorf(
			"making inode `%s` in dir `%d`: %w",
			params.Name,
			params.DirIno,
			err,
		)
	}
	if dirInode.Mode.Type != FileTypeDir {
		return Inode{}, fmt.Errorf(
			"making inode `%s` in dir `%d`: %w",
			params.Name,
			params.DirIno,
			NotDirErr,
		)
	}

	newIno, err := AllocInode(fs)
	if err != nil {
		return Inode{}, fmt.Errorf(
			"making inode `%s` in dir `%d`: %w",
			params.Name,
			params.DirIno,
			err,
		)
	}

	newInode, err := MakeInode(fs, &InodeParams{
		DirInode: &dirInode,
		Ino:      newIno,
		Mode:     params.Mode,
		Attr:     params.Attr,
	})
	if err != nil {
		return Inode{}, fmt.Errorf(
			"making inode `%s` in dir `%d`: %w",
			params.Name,
			params.DirIno,
			err,
		)
	}

	if err := AddDirEntry(fs, &dirInode, &newInode, params.Name); err != nil {
		return Inode{}, fmt.Errorf(
			"making inode `%s` in dir `%d`: %w",
			params.Name,
			params.DirIno,
			err,
		)
	}

	return newInode, nil
}

type FreeSpace struct {
	Current, Next, Prev Byte
}

func AddDirEntry(fs *FileSystem, dir *Inode, entry *Inode, name string) error {
	if dir.Mode.Type != FileTypeDir {
		return fmt.Errorf("adding dir entry: not a directory: %d", dir.Ino)
	}
	entrySize := DirEntrySize(Byte(len(name)))
	var placeForEntry *FreeSpace
	var offset Byte
	var lastOffset Byte

	for offset < dir.Size {
		dirEntry, err := ReadDirEntry(fs, dir, offset)
		if err != nil {
			return fmt.Errorf(
				"adding entry `%d` into dir `%d` with name `%s`: %w",
				entry.Ino,
				dir.Ino,
				name,
				err,
			)
		}

		if dirEntry.Name == name {
			if dirEntry.Header.Ino == entry.Ino {
				return nil
			}

			newHeader := DirEntryHeader{
				Ino:      entry.Ino,
				RecLen:   dirEntry.Header.RecLen,
				NameLen:  dirEntry.Header.NameLen,
				FileType: entry.Mode.Type,
			}

			if err := WriteDirEntry(
				fs,
				dir,
				&DirEntry{
					Header:     newHeader,
					Name:       "",
					NextOffset: offset,
				},
			); err != nil {
				return fmt.Errorf(
					"adding entry `%d` into dir `%d` with name `%s`: %w",
					entry.Ino,
					dir.Ino,
					name,
					err,
				)
			}
			entry.LinksCount++

			oldInode, err := GetInode(fs, entry.Ino)
			if err != nil {
				return fmt.Errorf(
					"adding entry `%d` into dir `%d` with name `%s`: %w",
					entry.Ino,
					dir.Ino,
					name,
					err,
				)
			}

			if err := UnlinkInode(fs, &oldInode); err != nil {
				return fmt.Errorf(
					"adding entry `%d` into dir `%d` with name `%s`: %w",
					entry.Ino,
					dir.Ino,
					name,
					err,
				)
			}
		}

		var freeOffset Byte
		if entry.Ino == InoOutOfInodes {
			freeOffset = offset
		} else {
			freeOffset = offset + DirEntrySize(Byte(len(dirEntry.Name)))
		}
		blockSize := fs.Superblock.BlockSize
		spaceInBlock := ((offset / blockSize * blockSize) + blockSize) - offset
		freeSize := Min(dirEntry.NextOffset-freeOffset, spaceInBlock)

		if placeForEntry == nil && freeSize >= entrySize {
			placeForEntry = &FreeSpace{
				Current: freeOffset,
				Prev:    offset,
				Next:    dirEntry.NextOffset,
			}
		}

		lastOffset = offset
		offset = dirEntry.NextOffset
	}

	if err := InsertDirEntry(
		fs,
		&DirEntryInsertParams{
			Dir:           dir,
			Entry:         entry,
			Name:          name,
			PlaceForEntry: placeForEntry,
			LastOffset:    lastOffset,
		},
	); err != nil {
		return fmt.Errorf(
			"adding entry `%d` into dir `%d` with name `%s`: %w",
			entry.Ino,
			dir.Ino,
			name,
			err,
		)
	}

	return nil
}

type DirEntryInsertParams struct {
	Dir           *Inode
	Entry         *Inode
	Name          string
	PlaceForEntry *FreeSpace
	LastOffset    Byte
}

const (
	SizeMaxFileName int = 256
)

func InsertDirEntry(fs *FileSystem, params *DirEntryInsertParams) error {
	// TODO: Can we prevent this upstream with types? E.g., a custom FileName
	// type with a constructor that validates the size requirement?
	if len(params.Name) >= SizeMaxFileName {
		return fmt.Errorf(
			"inserting entry `%d` into dir `%d` with name `%s`: %w",
			params.Entry.Ino,
			params.Dir.Ino,
			params.Name,
			FileNameTooLongErr,
		)
	}

	if params.PlaceForEntry == nil {
		blockSize := fs.Superblock.BlockSize
		nextBlock := params.LastOffset/blockSize + 1
		params.PlaceForEntry = &FreeSpace{
			Current: nextBlock + blockSize,
			Prev:    params.LastOffset,
			Next:    (nextBlock + 1) * blockSize,
		}
	}

	newHeader := DirEntryHeader{
		Ino:      params.Entry.Ino,
		RecLen:   (params.PlaceForEntry.Next - params.PlaceForEntry.Current),
		NameLen:  byte(len(params.Name)),
		FileType: params.Entry.Mode.Type,
	}
	if err := WriteDirEntry(
		fs,
		params.Dir,
		&DirEntry{
			Header:     newHeader,
			Name:       params.Name,
			NextOffset: params.PlaceForEntry.Current,
		},
	); err != nil {
		return fmt.Errorf(
			"inserting entry `%d` into dir `%d` with name `%s`: %w",
			params.Entry.Ino,
			params.Dir.Ino,
			params.Name,
			err,
		)
	}
	if err := WriteDirEntryRecLen(
		fs,
		params.Dir,
		params.PlaceForEntry.Prev,
		params.PlaceForEntry.Current-params.PlaceForEntry.Prev,
	); err != nil {
		return fmt.Errorf(
			"inserting entry `%d` into dir `%d` with name `%s`: %w",
			params.Entry.Ino,
			params.Dir.Ino,
			params.Name,
			err,
		)
	}
	params.Entry.LinksCount++
	if err := UpdateInode(fs, params.Entry); err != nil {
		return fmt.Errorf(
			"inserting entry `%d` into dir `%d` with name `%s`: %w",
			params.Entry.Ino,
			params.Dir.Ino,
			params.Name,
			err,
		)
	}

	return nil
}

func WriteDirEntryRecLen(
	fs *FileSystem,
	dir *Inode,
	offset Byte,
	recLen Byte,
) error {
	var p [SizeByte]byte
	putByte(p[:], recLen)
	if _, err := WriteInodeData(
		fs,
		dir,
		offset+dirEntryFieldRecLenOffset,
		p[:],
	); err != nil {
		return fmt.Errorf(
			"writing entry record length for dir `%d`: %w",
			dir.Ino,
			err,
		)
	}
	return nil
}

type DirEntry struct {
	Header     DirEntryHeader
	Name       string
	NextOffset Byte
}

func ReadDirEntry(
	fs *FileSystem,
	inode *Inode,
	offset Byte,
) (DirEntry, error) {
	buf := make([]byte, DirEntrySize(0))
	offset, err := ReadInodeData(fs, inode, offset, buf)
	if err != nil {
		return DirEntry{}, fmt.Errorf(
			"reading dir entry for inode `%d` at offset `%d`: %w",
			inode.Ino,
			offset,
			err,
		)
	}
	var header DirEntryHeader
	DecodeDirEntryHeader(&header, (*[SizeDirEntryHeader]byte)(buf))

	if header.RecLen < DirEntrySize(Byte(header.NameLen)) {
		return DirEntry{}, fmt.Errorf(
			"reading dir entry for inode `%d` at offset `%d`: %w",
			inode.Ino,
			offset,
			EntryTooShortErr,
		)
	}

	nameBuffer := make([]byte, header.NameLen)
	if _, err := ReadInodeData(
		fs,
		inode,
		DirEntrySize(offset),
		nameBuffer,
	); err != nil {
		return DirEntry{}, fmt.Errorf(
			"reading dir entry for inode `%d` at offset `%d`: %w",
			inode.Ino,
			offset,
			err,
		)
	}

	return DirEntry{
		Header:     header,
		Name:       string(nameBuffer),
		NextOffset: offset + header.RecLen,
	}, nil
}

func MakeDir(fs *FileSystem, parent, dir *Inode) error {
	dotDotOffset := align4(DirEntrySize(1))
	buf := make([]byte, fs.Superblock.BlockSize)

	dotEntry := DirEntryHeader{
		Ino:      dir.Ino,
		RecLen:   dotDotOffset,
		NameLen:  1,
		FileType: FileTypeDir,
	}

	dotDotEntry := DirEntryHeader{
		Ino:      parent.Ino,
		RecLen:   fs.Superblock.BlockSize - dotDotOffset,
		NameLen:  2,
		FileType: FileTypeDir,
	}

	dotBuf := (*[SizeDirEntryHeader]byte)(buf[:SizeDirEntryHeader])
	dotDotBuf := (*[SizeDirEntryHeader]byte)(buf[dotDotOffset : dotDotOffset+SizeDirEntryHeader])
	EncodeDirEntryHeader(&dotEntry, dotBuf)
	EncodeDirEntryHeader(&dotDotEntry, dotDotBuf)

	buf[DirEntrySize(0)] = '.'
	buf[dotDotOffset+DirEntrySize(0)] = '.'
	buf[dotDotOffset+DirEntrySize(0)+1] = '.'

	log.Printf("writing inode data")
	if _, err := WriteInodeData(fs, dir, 0, buf); err != nil {
		return fmt.Errorf(
			"initializing inode `%d` as dir in `%d`: %w",
			dir.Ino,
			parent.Ino,
			err,
		)
	}

	parent.LinksCount++
	log.Printf("updating parent inode")
	if err := UpdateInode(fs, parent); err != nil {
		return fmt.Errorf(
			"initializing inode `%d` as dir in `%d`: %w",
			dir.Ino,
			parent.Ino,
			err,
		)
	}
	dir.LinksCount++
	log.Printf("updating directory inode")
	if err := UpdateInode(fs, dir); err != nil {
		return fmt.Errorf(
			"initializing inode `%d` as dir in `%d`: %w",
			dir.Ino,
			parent.Ino,
			err,
		)
	}

	fs.Descriptor.UsedDirsCount++
	// TODO: Mark descriptor dirty
	return nil
}

func DestroyDir(fs *FileSystem, dir *Inode) error {
	var dotIno, dotDotIno = InoOutOfInodes, InoOutOfInodes
	var offset Byte
	for offset < dir.Size {
		entry, err := ReadDirEntry(fs, dir, offset)
		if err != nil {
			return fmt.Errorf("destroying dir `%d`: %w", dir.Ino, err)
		}

		if entry.Header.Ino != InoOutOfInodes {
			if entry.Name == "." {
				dotIno = entry.Header.Ino
			} else if entry.Name == ".." {
				dotDotIno = entry.Header.Ino
			} else {
				return fmt.Errorf(
					"destroying dir `%d`: %w",
					dir.Ino,
					DirNotEmptyErr,
				)
			}
		}

		entry.Header.Ino = InoOutOfInodes
		if err := WriteDirEntry(
			fs,
			dir,
			&DirEntry{
				Header:     entry.Header,
				NextOffset: offset,
				Name:       "",
			},
		); err != nil {
			return fmt.Errorf("destroying dir `%d`: %w", dir.Ino, err)
		}
		offset = entry.NextOffset
	}

	if dotIno == InoOutOfInodes {
		return fmt.Errorf(
			"destroying dir `%d`: %w",
			dir.Ino,
			DirMissingSelfEntryErr,
		)
	}
	if dotIno == dir.Ino {
		dir.LinksCount--
	} else {
		return fmt.Errorf(
			"destroying dir `%d`: entry '.' points to `%d` instead of "+
				"`%d`",
			dir.Ino,
			dotIno,
			dir.Ino,
		)
	}

	if dotDotIno == InoOutOfInodes {
		return fmt.Errorf(
			"destroying dir `%d`: %w",
			dir.Ino,
			DirMissingParentEntryErr,
		)
	}
	if dotDotIno == dir.Ino {
		return fmt.Errorf(
			"destroying dir `%d`: %w",
			dir.Ino,
			DirIsOwnParentErr,
		)
	}
	parent, err := GetInode(fs, dotDotIno)
	if err != nil {
		return fmt.Errorf(
			"destroying dir `%d`: fetching parent inode: %w",
			dir.Ino,
			err,
		)
	}
	parent.LinksCount--
	if err := UpdateInode(fs, &parent); err != nil {
		return fmt.Errorf(
			"destroying dir `%d`: decrementing parent's link counter: %w",
			dir.Ino,
			err,
		)
	}

	fs.Descriptor.UsedDirsCount--
	// TODO: mark superblock dirty
	return nil
}

func IsDirEmpty(fs *FileSystem, dir *Inode) (bool, error) {
	var offset Byte
	for offset < dir.Size {
		entry, err := ReadDirEntry(fs, dir, offset)
		if err != nil {
			return false, fmt.Errorf(
				"checking dir `%d` emptiness: %w",
				dir.Ino,
				err,
			)
		}

		if entry.Header.Ino != InoOutOfInodes {
			if entry.Name != "." && entry.Name != ".." {
				return false, nil
			}
		}
		offset = entry.NextOffset
	}
	return true, nil
}

func align4(x Byte) Byte {
	return (x + 0b11) &^ 0b11
}
