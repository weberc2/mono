package main

import (
	"log"

	"github.com/weberc2/mono/ext2/pkg/ext2"
)

func main() {
	const volumeSize = 1024 * 1024
	volume := ext2.NewMemoryVolume(volumeSize)
	fs := ext2.NewFileSystem(
		&ext2.Superblock{
			BlocksCount:     1024,
			FreeBlocksCount: 990,
			FreeInodesCount: 117,
			FirstDataBlock:  1,
			LogBlockSize:    0,
			BlocksPerGroup:  8192,
			InodesPerGroup:  128,
			State:           ext2.StateClean,
			RevLevel:        ext2.RevLevelDynamic,
			FirstIno:        11,
			InodeSize:       128,
			FeatureCompat:   0,
			FeatureIncompat: ext2.SupportedIncompatFeatures,
			FeatureROCompat: ext2.SupportedROCompatFeatures,
		},
		volume,
	)

	if err := fs.Flush(); err != nil {
		log.Fatalf("initializing filesystem: %v", err)
	}
}
