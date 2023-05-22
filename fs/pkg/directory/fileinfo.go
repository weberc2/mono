package directory

import (
	. "github.com/weberc2/mono/fs/pkg/types"
)

type FileInfo struct {
	Ino      Ino
	FileType FileType
	Name     string
}

func (fi *FileInfo) Equal(other *FileInfo) bool {
	return fi.Ino == other.Ino && fi.FileType == other.FileType &&
		fi.Name == other.Name
}
