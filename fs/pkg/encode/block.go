package encode

import (
	"encoding/binary"

	. "github.com/weberc2/mono/fs/pkg/types"
)

func EncodeBlock(b Block, p *[BlockPointerSize]byte) {
	binary.LittleEndian.PutUint64((*p)[:], uint64(b))
}

func DecodeBlock(p *[BlockPointerSize]byte) Block {
	return Block(binary.LittleEndian.Uint64((*p)[:]))
}
