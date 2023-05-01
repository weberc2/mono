package store

import (
	"fmt"

	"github.com/weberc2/mono/fs/pkg/encode"
	"github.com/weberc2/mono/fs/pkg/io"
	. "github.com/weberc2/mono/fs/pkg/types"
)

type VolumeInodeStore struct {
	volume io.Volume
}

func NewVolumeInodeStore(volume io.Volume) VolumeInodeStore {
	return VolumeInodeStore{volume}
}

func (store VolumeInodeStore) Put(inode *Inode) error {
	buf := new([InodeSize]byte)
	encode.EncodeInode(inode, buf)
	offset := Byte(inode.Ino) * InodeSize
	if err := store.volume.WriteAt(offset, buf[:]); err != nil {
		return fmt.Errorf(
			"writing inode `%d` to volume at offset `%d`: %w",
			inode.Ino,
			offset,
			err,
		)
	}
	return nil
}

func (store VolumeInodeStore) Get(ino Ino, output *Inode) error {
	buf := new([InodeSize]byte)
	offset := Byte(ino) * InodeSize
	if err := store.volume.ReadAt(offset, buf[:]); err != nil {
		return fmt.Errorf(
			"reading inode `%d` from volume at offset `%d`: %w",
			ino,
			offset,
			err,
		)
	}
	encode.DecodeInode(output, buf)
	output.Ino = ino
	return nil
}
