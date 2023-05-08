package alloc

import (
	"github.com/weberc2/mono/fs/pkg/math"
	. "github.com/weberc2/mono/fs/pkg/types"
)

const bitsPerByte = 8

type Bitmap struct {
	bytes []byte
}

func New(size Byte) Bitmap {
	return Bitmap{make([]byte, math.DivRoundUp(size, bitsPerByte))}
}

func (bm Bitmap) Alloc() (uint64, bool) {
	i, bit, ok := bytesFirstZero(bm.bytes)
	if !ok {
		return 0, false
	}
	bm.bytes[i] = byteSetHigh(bm.bytes[i], bit)
	return uint64(i*bitsPerByte) + uint64(bit), true
}

func (bm Bitmap) Free(value uint64) {
	b := &bm.bytes[value/bitsPerByte]
	*b = byteSetLow(*b, uint8(value%bitsPerByte))
}

func (bm Bitmap) Reserve(value uint64) {
	b := &bm.bytes[value/bitsPerByte]
	*b = byteSetHigh(*b, uint8(value%bitsPerByte))
}

func (bm Bitmap) Bytes() []byte { return bm.bytes }

func bytesFirstZero(bytes []byte) (int, uint8, bool) {
	for i, byt := range bytes {
		if bit := byteFirstZero(byt); bit != 0xff {
			return i, bit, true
		}
	}
	return 0, 0, false
}

func byteIsZero(byt byte, bit uint8) bool {
	return byt&(0b1000_0000>>bit) == 0
}

func byteSetHigh(byt byte, bit uint8) byte {
	return byt | (0b1000_0000 >> bit)
}

func byteSetLow(byt byte, bit uint8) byte {
	return byt & ^(0b1000_0000 >> bit)
}

func byteFirstZero(byt byte) uint8 {
	for bit := uint8(0); bit < 8; bit++ {
		if byteIsZero(byt, bit) {
			return bit
		}
	}
	return 0xFF
}
