package types

type DirEntry struct {
	Ino      Ino
	FileType FileType
	NameLen  uint8
	RecLen   uint16
	Name     string
}
