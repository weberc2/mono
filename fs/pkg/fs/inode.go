package fs

import (
	"fmt"
	"time"
)

type Ino uint64

const (
	InoOutOfInodes Ino = iota

	// InoRoot is the root ino
	InoRoot

	// InoFirst is the first non-reserved ino
	InoFirst

	DirNotEmptyErr constErr = "directory not empty"
)

func (ino Ino) TableOffset() Byte { return Byte(ino) * InodeSize }

type Inode struct {
	Ino                 Ino
	Mode                Mode
	Attr                FileAttr
	Size                Byte
	LinksCount          uint16
	Flags               uint32
	ACL                 uint32
	DirectBlocks        [DirectBlocksPerInode]Block
	SinglyIndirectBlock Block
	DoublyIndirectBlock Block
	TriplyIndirectBlock Block
}

type Mode struct {
	Type         FileType
	SUID         bool
	SGID         bool
	Sticky       bool
	AccessRights uint16
}

type FileAttr struct {
	UID   uint32
	GID   uint32
	ATime uint32
	CTime uint32
	MTime uint32
	DTime uint32
}

type FileType byte

const (
	FileTypeRegular FileType = iota
	FileTypeDir
	FileTypeCharDev
	FileTypeBlockDev
	FileTypeFifo
	FileTypeSocket
	FileTypeSymlink
	FileTypeNone
)

type DirEntryHeader struct {
	Ino      Ino
	RecLen   Byte
	NameLen  byte
	FileType FileType
}

// DirEntrySize returns the size of a directory entry with a name of `nameLen`
// bytes in length.
func DirEntrySize(nameLen Byte) Byte { return SizeDirEntryHeader + nameLen }

// GetInode fetches a particular ino. If the ino is in the inode cache, the
// cached copy will be returned, otherwise a new inode will be loaded from
// memory and pushed onto the cache, which may evict the least recently used
// inode--if an inode is evicted, this function will ensure that it is written
// to the volume.
func GetInode(fs *FileSystem, ino Ino) (Inode, error) {
	// if the ino is in the cache, return the cached copy
	if inode, ok := fs.InodeCache.Get(ino); ok {
		return inode, nil
	}

	// if it's not in the cache then we create a new one and cache it.
	var inode Inode
	if err := ReadInode(fs, ino, &inode); err != nil {
		return inode, err
	}

	// if pushing the new inode evicts another, write the evicted inode to
	// the volume
	if evicted := fs.InodeCache.Push(inode); evicted != nil {
		// TODO: Only write the evicted inode if it's dirty
		if err := WriteInode(fs, evicted); err != nil {
			return inode, err
		}
	}
	return inode, nil
}

func FlushIno(fs *FileSystem, ino Ino) error {
	inode, found := fs.InodeCache.Remove(ino)
	if !found {
		return nil
	}
	if err := WriteInode(fs, &inode); err != nil {
		return fmt.Errorf("flushing ino `%d`: %w", ino, err)
	}
	return nil
}

type InodeParams struct {
	DirInode *Inode
	Ino      Ino
	Mode     Mode
	Attr     FileAttr
}

func MakeInode(fs *FileSystem, params *InodeParams) (Inode, error) {
	inode := Inode{Ino: params.Ino, Mode: params.Mode, Attr: params.Attr}
	if params.Mode.Type == FileTypeDir {
		if err := MakeDir(fs, params.DirInode, &inode); err != nil {
			return Inode{}, fmt.Errorf(
				"creating inode `%d`: %w",
				params.Ino,
				err,
			)
		}
	}

	if err := UpdateInode(fs, &inode); err != nil {
		return Inode{}, fmt.Errorf("creating inode `%d`: %w", params.Ino, err)
	}

	return inode, nil
}

// UpdateInode updates the inode in the cache. As a side effect, if the cache
// is full, it will evict the least-recently-used node and flush it to disk.
func UpdateInode(fs *FileSystem, inode *Inode) error {
	fs.DirtyInos.Add(inode.Ino)
	if evicted := fs.InodeCache.Push(*inode); evicted != nil {
		// TODO: Only write the evicted inode if it's dirty
		if err := WriteInode(fs, evicted); err != nil {
			return err
		}
	}
	return nil
}

func UnlinkInode(fs *FileSystem, inode *Inode) error {
	if inode.Mode.Type == FileTypeDir {
		empty, err := IsDirEmpty(fs, inode)
		if err != nil {
			return fmt.Errorf("unlinking inode `%d`: %w", inode.Ino, err)
		}
		if !empty {
			return fmt.Errorf(
				"unlinking inode `%d`: %w",
				inode.Ino,
				DirNotEmptyErr,
			)
		}

		if inode.LinksCount != 2 {
			return fmt.Errorf(
				"unlinking inode `%d`: empty directory should have `2` "+
					"links, but has `%d`",
				inode.Ino,
				inode.LinksCount,
			)
		}

		if err := DestroyDir(fs, inode); err != nil {
			return fmt.Errorf("unlinking inode `%d`: %w", inode.Ino, err)
		}
	}

	inode.LinksCount--
	if inode.LinksCount == 0 {
		if err := RemoveInode(fs, inode); err != nil {
			return fmt.Errorf("unlinking inode `%d`: %w", inode.Ino, err)
		}
	}

	if err := UpdateInode(fs, inode); err != nil {
		return fmt.Errorf("unlinking inode `%d`: %w", inode.Ino, err)
	}
	return nil
}

func RemoveInode(fs *FileSystem, inode *Inode) error {
	if err := DeallocInodeBlocks(fs, inode); err != nil {
		return fmt.Errorf("removing inode `%d`: %w", inode.Ino, err)
	}
	inode.Attr.DTime = timestamp(time.Now())
	DeallocInode(fs, inode.Ino)
	return nil
}

func DeallocInodeBlocks(fs *FileSystem, inode *Inode) error {
	if !IsFastSymlink(fs, inode) {
		for _, block := range inode.DirectBlocks {
			DeallocInodeBlock(fs, inode, block)
		}
		DeallocIndirectBlock(fs, inode, inode.SinglyIndirectBlock, 1)
		DeallocIndirectBlock(fs, inode, inode.DoublyIndirectBlock, 2)
		DeallocIndirectBlock(fs, inode, inode.TriplyIndirectBlock, 3)
	}
	return nil
}

func DeallocIndirectBlock(
	fs *FileSystem,
	inode *Inode,
	indirectBlock Block,
	indirection InodeBlockDirection,
) error {
	if indirectBlock == BlockOutOfRange {
		return nil
	}

	// validate indirection
	switch indirection {
	case InodeBlockSinglyIndirect:
		break
	case InodeBlockDoublyIndirect:
		break
	case InodeBlockTriplyIndirect:
		break
	default:
		panic(fmt.Sprintf(
			"`indirection` must be one of `InodeBlockSinglyIndirect` (`%d`), "+
				"`InodeBlockDoublyIndirect` (`%d`), or "+
				"`InodeBlockTriplyIndirect` (`%d`); found `%d`",
			InodeBlockSinglyIndirect,
			InodeBlockDoublyIndirect,
			InodeBlockTriplyIndirect,
			indirection,
		))
	}

	blockSize := fs.Superblock.BlockSize
	p := make([]byte, blockSize)
	offset := Byte(indirectBlock) * blockSize
	if err := ReadAt(fs.Volume, offset, p); err != nil {
		return fmt.Errorf(
			"deallocating indirect block `%d` in inode `%d`: %w",
			indirectBlock,
			inode.Ino,
			err,
		)
	}
	for i := Byte(0); i < blockSize; i += BlockPointerSize {
		block := getBlock(p[i:])
		if block != 0 && indirection > InodeBlockSinglyIndirect {
			if err := DeallocIndirectBlock(
				fs,
				inode,
				block,
				indirection-1,
			); err != nil {
				return fmt.Errorf(
					"deallocating indirect block `%d` in inode `%d`: %w",
					indirectBlock,
					inode.Ino,
					err,
				)
			}
		} else if block != 0 {
			DeallocInodeBlock(fs, inode, block)
		}
	}
	inode.Size -= blockSize
	DeallocBlock(fs, indirectBlock)
	return nil
}

func DeallocInodeBlock(fs *FileSystem, inode *Inode, block Block) {
	if block == BlockOutOfRange {
		return
	}
	inode.Size -= fs.Superblock.BlockSize
	DeallocBlock(fs, block)
}

type inodeField int

const (
	size64 uint16 = 8
	size32 uint16 = 4
	size16 uint16 = 2
)

const (
	inodeFieldAttrUID inodeField = iota
	inodeFieldAttrGID
	inodeFieldAttrATime
	inodeFieldAttrCTime
	inodeFieldAttrMTime
	inodeFieldAttrDTime
	inodeFieldFlags
	inodeFieldACL
	inodeFieldMode
	inodeFieldLinksCount
	inodeFieldPad0
	inodeFieldSize
	inodeFieldDirectBlocks
	inodeFieldSinglyIndirectBlock
	inodeFieldDoublyIndirectBlock
	inodeFieldTriplyIndirectBlock
	inodeFieldEOF
)

var (
	inodeFieldSizes = [inodeFieldEOF]uint16{
		inodeFieldAttrUID:             size32, // first 64-bits
		inodeFieldAttrGID:             size32,
		inodeFieldAttrATime:           size32, // second 64-bits
		inodeFieldAttrCTime:           size32,
		inodeFieldAttrMTime:           size32, // third 64-bits
		inodeFieldAttrDTime:           size32,
		inodeFieldFlags:               size32, // fourth 64-bits
		inodeFieldACL:                 size32,
		inodeFieldMode:                size16, // fifth 64-bits
		inodeFieldLinksCount:          size16,
		inodeFieldPad0:                size32,
		inodeFieldSize:                size64, // rest
		inodeFieldDirectBlocks:        uint16(DirectBlocksPerInode * 64),
		inodeFieldSinglyIndirectBlock: size64,
		inodeFieldDoublyIndirectBlock: size64,
		inodeFieldTriplyIndirectBlock: size64,
	}

	inodeFieldOffsets = [inodeFieldEOF]uint16{} // generated in init
)

func init() {
	var lastOffset uint16 = 0
	for field := inodeField(0); field < inodeFieldEOF; field++ {
		inodeFieldOffsets[field] = lastOffset
		lastOffset += inodeFieldSizes[field]
	}
}
