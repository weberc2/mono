package fs

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

type Descriptor struct {
	UsedDirsCount uint64
	BlockBitmap   Bitmap
	InodeBitmap   Bitmap
}

func NewDescriptor(blocks Block, inodes Ino) Descriptor {
	return Descriptor{
		UsedDirsCount: 0,
		BlockBitmap:   make(Bitmap, DivCiel(blocks, 8)),
		InodeBitmap:   make(Bitmap, DivCiel(inodes, 8)),
	}
}

func (d *Descriptor) Debug() string {
	debugDescriptor := struct {
		UsedDirsCount uint64
		BlockBitmap   string
		InodeBitmap   string
	}{
		UsedDirsCount: d.UsedDirsCount,
		BlockBitmap:   "0x" + hex.EncodeToString(d.BlockBitmap),
		InodeBitmap:   "0x" + hex.EncodeToString(d.BlockBitmap),
	}
	data, err := json.Marshal(&debugDescriptor)
	if err != nil {
		panic(fmt.Sprintf(
			"ERROR failed to marshal debug descriptor `%#v` to JSON: %v",
			debugDescriptor,
			err,
		))
	}
	return string(data)
}

func (d *Descriptor) Equal(other *Descriptor) bool {
	return d.UsedDirsCount == other.UsedDirsCount &&
		bytes.Equal(d.BlockBitmap, other.BlockBitmap) &&
		bytes.Equal(d.InodeBitmap, other.InodeBitmap)
}
