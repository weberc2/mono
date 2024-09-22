package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/cenkalti/rain/rainrpc"
)

func RunFromPath(ctx Context, configPath string) error {
	app, err := LoadAppFromPath(configPath)
	if err != nil {
		return err
	}

	return app.Run(ctx)
}

type Config struct {
	Addr               string `json:"addr"`
	RainServerURL      string `json:"rainServerURL"`
	DownloadsDirectory string `json:"downloadsDirectory"`
	FilmsDirectory     string `json:"filmsDirectory"`
	ShowsDirectory     string `json:"showsDirectory"`
}

func LoadConfigFromPath(path string) (config Config, err error) {
	var data []byte
	if data, err = os.ReadFile(path); err != nil {
		err = fmt.Errorf("loading config file from path: %w", err)
		return
	}

	if err = json.Unmarshal(data, &config); err != nil {
		err = fmt.Errorf("loading config file from path: %w", err)
	}
	return
}

func LoadAppFromPath(path string) (app App, err error) {
	var config Config
	if config, err = LoadConfigFromPath(path); err != nil {
		err = fmt.Errorf("loading app: %w", err)
		return
	}

	downloads := newDownloadStore()
	app.DownloadController = DownloadController{
		DownloadDirectory: config.DownloadsDirectory,
		FilmsDirectory:    config.FilmsDirectory,
		ShowsDirectory:    config.ShowsDirectory,
		Downloads:         downloads,
		Torrents:          rainrpc.NewClient(config.RainServerURL),
	}

	var mux http.ServeMux
	(&API{Downloads: downloads}).Register(&mux)
	app.Server = http.Server{Addr: config.Addr, Handler: &mux}

	return
}
