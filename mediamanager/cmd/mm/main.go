package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"mediamanager/pkg/mm"
	"mediamanager/pkg/mm/api"
	"time"

	"github.com/cenkalti/rain/rainrpc"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if err := Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func Run(ctx context.Context) error {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	conn, err := pgxpool.New(ctx, "")
	if err != nil {
		return fmt.Errorf("connecting to postgres database: %w", err)
	}
	defer conn.Close()

	downloads := mm.PostgresDownloadStore{DB: conn}
	imports := mm.PostgresImportStore{DB: conn}

	api := api.API{
		Downloads: downloads,
		Imports:   &imports,
		Logger:    slog.Default().With("component", "DOWNLOAD-API"),
	}

	importController := mm.ImportController{
		Downloads: downloads,
		Imports:   &imports,
		Logger:    slog.Default().With("component", "IMPORT-CONTROLLER"),
		Importer: mm.Importer{
			DownloadsDirectory: "/rain/data",
			FilmsDirectory:     "/rain/data/Movies",
			ScratchDirectory:   "/rain/data/.scratch",
		},
	}

	downloadController := mm.DownloadController{
		Torrents:  rainrpc.NewClient("http://rain:7246"),
		Downloads: &downloads,
		Logger:    slog.Default().With("component", "DOWNLOAD-CONTROLLER"),
		Trackers: []string{
			"udp://open.demonii.com:1337",
			"udp://tracker.coppersurfer.tk:6969",
			"udp://tracker.leechers-paradise.org:6969",
			"udp://tracker.pomf.se:80",
			"udp://tracker.publicbt.com:80",
			"udp://tracker.openbittorrent.com:80",
			"udp://tracker.istole.it:80",
			"udp://explodie.org:6969",
			"udp://tracker.empire-js.us:1337",
			"udp://tracker.opentrackr.org:1337",
			"http://tracker.opentrackr.org:1337/announce",
			"udp://open.stealth.si:80/announce",
			"udp://tracker.torrent.eu.org:451/announce",
			"udp://explodie.org:6969/announce",
			"udp://exodus.desync.com:6969/announce",
			"udp://tracker.0x7c0.com:6969/announce",
			"udp://tracker-udp.gbitt.info:80/announce",
			"udp://retracker01-msk-virt.corbina.net:80/announce",
			"udp://opentracker.io:6969/announce",
			"udp://moonburrow.club:6969/announce",
			"udp://bt.ktrackers.com:6666/announce",
			"https://tracker.tamersunion.org:443/announce",
			"http://tracker1.bt.moack.co.kr:80/announce",
			"udp://tracker2.dler.org:80/announce",
			"udp://ryjer.com:6969/announce",
			"udp://run.publictracker.xyz:6969/announce",
			"udp://open.dstud.io:6969/announce",
			"udp://bt2.archive.org:6969/announce",
			"udp://bt1.archive.org:6969/announce",
		},
	}

	results := make(chan error)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() { results <- importController.Run(ctx, 4*time.Second) }()
	go func() { results <- downloadController.Run(ctx, 4*time.Second) }()
	go func() { results <- api.Run(ctx, "0.0.0.0:8080") }()

	for range 3 {
		if err := <-results; err != nil {
			return err
		}
	}

	return nil
}
