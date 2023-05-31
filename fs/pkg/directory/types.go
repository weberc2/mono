package directory

import (
	. "github.com/weberc2/mono/fs/pkg/types"
)

type FreeSpace struct {
	Offset     Byte
	PrevOffset Byte
	NextOffset Byte
}

const (
	NameTooLongErr ConstError = "name too long"
	NotADirErr     ConstError = "not a directory"
)
