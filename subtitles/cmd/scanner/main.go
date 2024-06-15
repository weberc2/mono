package main

import (
	"log"
	"os"
	"subtitles/pkg/subtitles"
)

func main() {
	scanner := subtitles.Scanner{
		FileSystem:     os.DirFS("/tmp/media"),
		ShowsDirectory: "TV",
	}

	if err := scanner.Scan(); err != nil {
		log.Fatal(err)
	}
}
