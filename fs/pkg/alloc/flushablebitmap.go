package alloc

import (
	"sync"

	. "github.com/weberc2/mono/fs/pkg/types"
)

type BitmapStore interface {
	Put(Bitmap) error
}

type FlushableBitmap struct {
	bitmap Bitmap
	store  BitmapStore
	mutex  sync.Mutex
	dirty  bool
}

func NewFlushable(size Byte, store BitmapStore) *FlushableBitmap {
	return &FlushableBitmap{bitmap: New(size), store: store}
}

func (bitmap *FlushableBitmap) Alloc() (uint64, bool) {
	bitmap.mutex.Lock()
	defer bitmap.mutex.Unlock()
	bitmap.dirty = true
	return bitmap.bitmap.Alloc()
}

func (bitmap *FlushableBitmap) Reserve(handle uint64) {
	bitmap.mutex.Lock()
	defer bitmap.mutex.Unlock()
	bitmap.bitmap.Reserve(handle)
	bitmap.dirty = true
}

func (bitmap *FlushableBitmap) Free(handle uint64) {
	bitmap.mutex.Lock()
	defer bitmap.mutex.Unlock()
	bitmap.bitmap.Free(handle)
	bitmap.dirty = true
}

func (bitmap *FlushableBitmap) Flush() error {
	bitmap.mutex.Lock()
	defer bitmap.mutex.Unlock()
	if bitmap.dirty {
		if err := bitmap.store.Put(bitmap.bitmap); err != nil {
			return err
		}
		bitmap.dirty = false
	}
	return nil
}
