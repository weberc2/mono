package types

type Block uint64

const (
	BlockSize        Byte = 1024
	BlockPointerSize Byte = 8

	BlockNil Block = 0
)
