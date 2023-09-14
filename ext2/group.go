package main

type Group struct {
	Idx         uint64
	Desc        GroupDesc
	BlockBitmap []byte
	InodeBitmap []byte
	Dirty       bool
}
