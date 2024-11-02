package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Addr               string `json:"addr"`
	RainServerURL      string `json:"rainServerURL"`
	DownloadsDirectory string `json:"downloadsDirectory"`
	FilmsDirectory     string `json:"filmsDirectory"`
	ScratchDirectory   string `json:"scratchDirectory"`
	ShowsDirectory     string `json:"showsDirectory"`
}

func LoadConfig(data []byte) (config Config, err error) {
	if err = json.Unmarshal(data, &config); err != nil {
		err = fmt.Errorf("loading config: %w", err)
	}
	return
}

func LoadConfigFile(path string) (config Config, err error) {
	var data []byte
	if data, err = os.ReadFile(path); err != nil {
		err = fmt.Errorf("loading config file `%s`: %w", path, err)
		return
	}

	return LoadConfig(data)
}
