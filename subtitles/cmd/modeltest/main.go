package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	"subtitles/pkg/subtitles"
)

func main() {
	db, err := sql.Open("postgres", "")
	if err != nil {
		log.Fatalf("opening postgres: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	model := subtitles.Model{DB: db}
	episode := subtitles.Episode{
		Title:   "Test",
		Year:    "2024",
		Season:  "01",
		Episode: "01",
	}
	language := "en"
	if err := model.ForceDeleteDownload(ctx, &episode, language); err != nil {
		log.Fatal(err)
	}
	if _, err := model.InsertPendingEpisodeDownload(
		ctx,
		&episode,
		language,
	); err != nil {
		log.Fatal(err)
	}

	token, downloads, err := model.TryReserveDownloads(ctx, 1)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf(
		"token: %s; len(downloads): %d",
		time.Time(token),
		len(downloads),
	)

	token, downloads, err = model.TryReserveDownloads(ctx, 1)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf(
		"token: %s; len(downloads): %d",
		time.Time(token),
		len(downloads),
	)

	time.Sleep(10 * time.Second)

	token, downloads, err = model.TryReserveDownloads(ctx, 1)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf(
		"token: %s; len(downloads): %d",
		time.Time(token),
		len(downloads),
	)
}
