package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 1 {
		log.Fatalln("USAGE: ext2 FILE")
	}

	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalf("FATAL opening file volume: %v", err)
	}
	defer file.Close()

	volume := FileVolume{file}
	var buf [1024]byte
	if err := volume.Read(1024, buf[:]); err != nil {
		log.Fatalf("FATAL reading volume: %v", err)
	}

	superblock, err := DecodeSuperblock(&buf, false)
	if err != nil {
		log.Fatalf("FATAL decoding superblock: %v", err)
	}

	data, err := json.Marshal(superblock)
	if err != nil {
		log.Fatalf("FATAL marshaling superblock: %v", err)
	}

	fmt.Printf("%s\n", data)
}
