package fs

type Bit uint8

type Bitmap []byte

func (bitmap Bitmap) Len() Byte { return Byte(len(bitmap)) }

func (bitmap Bitmap) FirstZero() (Byte, Bit, bool) {
	for byt := range bitmap {
		for bit := Bit(0); bit < 8; bit++ {
			if (bitmap[byt]<<bit)&0b1000000 == 0 {
				return Byte(byt), bit, true
			}
		}
	}
	return 0, 0, false
}

func (bitmap Bitmap) SetHigh(byt Byte, bit Bit) {
	bitmap[byt] |= 1 << bit
}

func (bitmap Bitmap) SetLow(byt Byte, bit Bit) {
	bitmap[byt] &= ^(1 << bit)
}

func (bitmap Bitmap) Alloc() (uint64, bool) {
	byt, bit, ok := bitmap.FirstZero()
	if !ok {
		return 0, ok
	}
	bitmap.SetHigh(byt, bit)
	return uint64(byt*8) + uint64(bit), true
}

func (bitmap Bitmap) Free(i uint64) {
	byt := Byte(i / 8)
	bit := Bit(i % 8)
	bitmap.SetLow(byt, bit)
}
