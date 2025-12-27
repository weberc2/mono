package main

import (
	"dsmspaces/pkg/dsmspaces"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	slog.SetDefault(slog.New(slog.NewJSONHandler(
		os.Stderr,
		&slog.HandlerOptions{Level: slog.LevelDebug},
	)))

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}

	indexFile := os.Getenv("INDEXFILE")
	if indexFile == "" {
		return fmt.Errorf("missing required environment variable: INDEXFILE")
	}
	if _, err := os.Stat(indexFile); err != nil {
		return fmt.Errorf("checking index file: %w", err)
	}

	apiKey := os.Getenv("OPENAIAPIKEY")
	if apiKey == "" {
		return fmt.Errorf("missing required environment variable: OPENAIAPIKEY")
	}

	placesFile := os.Getenv("PLACESFILE")
	if placesFile == "" {
		return fmt.Errorf("missing required environment variable: PLACESFILE")
	}

	pmTilesFile := os.Getenv("PMTILESFILE")
	if pmTilesFile == "" {
		return fmt.Errorf("missing required environment variable: PMTILESFILE")
	}
	if _, err := os.Stat(pmTilesFile); err != nil {
		return fmt.Errorf("checking PMTiles file: %w", err)
	}

	data, err := os.ReadFile(placesFile)
	if err != nil {
		return fmt.Errorf("reading places file: %w", err)
	}

	var places []dsmspaces.Place
	if err := json.Unmarshal(data, &places); err != nil {
		return fmt.Errorf("unmarshaling places file `%s`: %w", placesFile, err)
	}

	slog.Info("starting http server", "addr", addr)
	return http.ListenAndServe(addr, &dsmspaces.Server{
		IndexFile:   indexFile,
		Places:      places,
		PMTilesFile: pmTilesFile,
		IntentParser: dsmspaces.NewIntentsParser(
			apiKey,
			dsmspaces.WithRecorder(
				dsmspaces.NewFileIntentsParseRecorder("./parses.jsonl"),
			),
		),
	})
}
