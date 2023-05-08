package store

import (
	"fmt"

	"github.com/weberc2/mono/fs/pkg/alloc"
	"github.com/weberc2/mono/fs/pkg/io"
)

var _ alloc.BitmapStore = VolumeBitmapStore{}

type VolumeBitmapStore struct {
	volume io.Volume
}

func NewVolumeBitmapStore(volume io.Volume) VolumeBitmapStore {
	return VolumeBitmapStore{volume}
}

func (store VolumeBitmapStore) Put(bitmap alloc.Bitmap) error {
	if err := store.volume.WriteAt(0, bitmap.Bytes()); err != nil {
		return fmt.Errorf("storing bitmap: %w", err)
	}
	return nil
}
