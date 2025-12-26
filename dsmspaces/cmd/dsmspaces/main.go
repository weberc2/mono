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
	indexFile := os.Getenv("INDEXFILE")
	if indexFile == "" {
		return fmt.Errorf("missing required environment variable: INDEXFILE")
	}

	apiKey := os.Getenv("OPENAIAPIKEY")
	if apiKey == "" {
		return fmt.Errorf("missing required environment variable: OPENAIAPIKEY")
	}

	placesFile := os.Getenv("PLACESFILE")
	if placesFile == "" {
		return fmt.Errorf("missing required environment variable: PLACESFILE")
	}

	data, err := os.ReadFile(placesFile)
	if err != nil {
		return fmt.Errorf("reading places file: %w", err)
	}

	var places []dsmspaces.Place
	if err := json.Unmarshal(data, &places); err != nil {
		return fmt.Errorf("unmarshaling places file `%s`: %w", placesFile, err)
	}

	server := dsmspaces.NewServer(indexFile, places, apiKey)
	const addr = ":8080"
	slog.Info("starting http server", "addr", addr)
	return http.ListenAndServe(addr, &server)
}
