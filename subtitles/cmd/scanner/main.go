package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"subtitles/pkg/subtitles"
)

func main() {
	db, err := sql.Open("postgres", "")
	if err != nil {
		log.Fatalf("opening database: %v", err)
	}
	defer db.Close()

	if len(os.Args) < 2 {
		log.Fatal("USAGE: subtitles <SHOWSDIR>")
	}

	scanner := subtitles.Scanner{
		Model:          subtitles.Model{DB: db},
		FileSystem:     os.DirFS(os.Args[1]),
		ShowsDirectory: ".",
	}

	if err := scanner.Scan(context.Background()); err != nil {
		log.Fatal(err)
	}
}
