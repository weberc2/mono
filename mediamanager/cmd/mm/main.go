package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/cenkalti/rain/rainrpc"
)

func main() {
	// slog.SetLogLoggerLevel(slog.LevelDebug)

	rain := rainrpc.NewClient("http://rain:7246")

	transforms := MemoryTransformStore{}

	downloader := NewDownloader(
		rain,
		4*time.Second,
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
	)
	controller := TransformController{
		Downloads:          &downloader,
		Transforms:         &transforms,
		DownloadsDirectory: "/rain/data",
		FilmsDirectory:     "/rain/data/Movies",
		Logger:             slog.With("component", "TRANSFORM-CONTROLLER"),
	}

	api := API{
		Logger:     slog.Default().With(slog.String("component", "api")),
		Transforms: &transforms,
		Downloads:  &downloader,
		Client:     http.DefaultClient,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		defer cancel()
		if err := api.Run(ctx, ":8080"); err != nil {
			slog.Error("running api", "err", err.Error())
			return
		}
	}()

	if err := controller.Run(ctx, 2*time.Second); err != nil {
		slog.Error("running controller", "err", err.Error())
		return
	}
}

type Context = context.Context
