package indirect

import (
	. "github.com/weberc2/mono/fs/pkg/types"
)

func offset(indirect Block, index Index) Byte {
	startOfBlock := Byte(indirect) * BlockSize
	offsetInBlock := Byte(index) * BlockPointerSize
	return startOfBlock + offsetInBlock
}
