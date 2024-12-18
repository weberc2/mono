package main

import (
	"dedup/pkg/dedup"
	"log"
	"os"
)

func main() {
	if err := dedup.Dedup(
		dedup.NewNotifier(os.Stdout),
		os.Args[1],
	); err != nil {
		log.Fatal(err)
	}
}
