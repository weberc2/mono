package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"subtitles/pkg/opensubtitles"
)

func main() {
	client := opensubtitles.Client{
		HTTP:      http.DefaultClient,
		APIKey:    "W2f56eccCa2o5hVnlKhcxAwA5DZgF1eY",
		UserAgent: "fetcher v0.0.0",
	}
	subtitles, err := client.Search(
		context.Background(),
		&opensubtitles.Query{
			Type:  opensubtitles.QueryTypeEpisode,
			Title: "The Office",
			// Year:      "2003",
			Season:    "07",
			Episode:   "07",
			Languages: "en",
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	data, err := json.Marshal(&subtitles)
	if err != nil {
		log.Fatal("marshaling subtitles", err)
	}
	fmt.Printf("%s\n", data)
}
