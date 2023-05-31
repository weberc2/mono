package ext2

type DynamicBitmap []byte

//		fn find_zero_bit_in_bitmap(bitmap: &[u8]) -> Option<(u64, u64)> {
//	   for byte in 0..bitmap.len() as u64 {
//	     if bitmap[byte as usize] == 0xff {
//	       continue
//	     }
//	     for bit in 0..8 {
//	       if (bitmap[byte as usize] & (1 << bit)) == 0 {
//	         return Some((byte, bit))
//	       }
//	     }
//	   }
//	   None
//		}
func (bitmap DynamicBitmap) FindZeroBit() (uint64, uint64, bool) {
	for byt := 0; byt < len(bitmap); byt++ {
		if bitmap[byt] == 0xff {
			continue
		}
		for bit := 0; bit < 8; bit++ {
			if (bitmap[byt] & (1 << bit)) == 0 {
				return uint64(byt), uint64(bit), true
			}
		}
	}
	return 0, 0, false
}

func (bitmap DynamicBitmap) FindZeroBitAfter(bit uint64) (uint64, uint64, bool) {
	byt := bit / 8
	if bitmap[byt] != 0xff {
		for bit := bit % 8; bit < 8; bit++ {
			if (bitmap[byt] & (1 << bit)) == 0 {
				return byt, bit, true
			}
		}
	}
	for byt := (bit / 8) + 1; byt < uint64(len(bitmap)); byt++ {
		if bitmap[byt] != 0xff {
			for bit := 0; bit < 8; bit++ {
				if (bitmap[byt] & (1 << bit)) == 0 {
					return byt, uint64(bit), true
				}
			}
		}
	}

	return 0, 0, false
}

func (bitmap DynamicBitmap) SetHigh(byt, bit uint64) {
	bitmap[byt] |= 1 << bit
}
