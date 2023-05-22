package encode

import (
	"encoding/binary"

	. "github.com/weberc2/mono/fs/pkg/types"
)

func putIno(b []byte, start Byte, u Ino) {
	putU64(b, start, uint64(u))
}

func getIno(b []byte, start Byte) Ino {
	return Ino(getU64(b, start))
}

func putBytePointer(b []byte, start Byte, u Byte) {
	putU64(b, start, uint64(u))
}

func getBytePointer(b []byte, start Byte) Byte {
	return Byte(getU64(b, start))
}

func putU64(b []byte, start Byte, u uint64) {
	binary.LittleEndian.PutUint64(b[start:start+8], u)
}

func getU64(b []byte, start Byte) uint64 {
	return binary.LittleEndian.Uint64(b[start : start+8])
}

func putU32(b []byte, start Byte, u uint32) {
	binary.LittleEndian.PutUint32(b[start:start+4], u)
}

func getU32(b []byte, start Byte) uint32 {
	return binary.LittleEndian.Uint32(b[start : start+4])
}

func putU16(b []byte, start Byte, u uint16) {
	binary.LittleEndian.PutUint16(b[start:start+2], u)
}

func getU16(b []byte, start Byte) uint16 {
	return binary.LittleEndian.Uint16(b[start : start+2])
}

func putU8(b []byte, start Byte, u uint8) {
	b[start] = u
}

func getU8(b []byte, start Byte) uint8 {
	return b[start]
}
