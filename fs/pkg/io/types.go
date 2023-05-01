package io

import (
	. "github.com/weberc2/mono/fs/pkg/types"
)

type ReadAt interface {
	ReadAt(offset Byte, b []byte) error
}

type WriteAt interface {
	WriteAt(offset Byte, p []byte) error
}

type Volume interface {
	ReadAt
	WriteAt
}
