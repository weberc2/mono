package main

type GroupDesc struct {
	BlockBitmap     uint32
	InodeBitmap     uint32
	InodeTable      uint32
	FreeBlocksCount uint16
	FreeInodesCount uint16
	UsedDirsCount   uint16
}
