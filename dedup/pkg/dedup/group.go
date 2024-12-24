package dedup

// Group is a collection of files with the same size, first, and final blocks.
type Group struct {
	// Size is the size of the files in the group.
	Size int64

	// FirstBlockChecksum is the checksum of the first block in the files.
	FirstBlockChecksum uint32

	// FinalBlockChecksum is the checksum of the final block in the files.
	FinalBlockChecksum uint32

	// Paths are the paths to the files in the group.
	Paths []string
}
