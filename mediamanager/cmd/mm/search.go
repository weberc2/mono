package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"
)

func Search(
	client *http.Client,
	ctx context.Context,
	query string,
) (results []SearchResult, err error) {
	var req *http.Request
	if req, err = http.NewRequestWithContext(
		ctx,
		"GET",
		"https://apibay.org/q.php?"+url.Values{"q": []string{query}}.Encode(),
		nil,
	); err != nil {
		err = fmt.Errorf(
			"searching for `%s`: building http request: %w",
			query,
			err,
		)
		return
	}

	var rsp *http.Response
	if rsp, err = client.Do(req); err != nil {
		err = fmt.Errorf(
			"searching for `%s`: sending http request: %w",
			query,
			err,
		)
		return
	}

	defer func() { err = errors.Join(err, rsp.Body.Close()) }()

	var data []byte
	const maxBytes = 1024 * 1024 * 1024
	if data, err = io.ReadAll(io.LimitReader(rsp.Body, maxBytes)); err != nil {
		err = fmt.Errorf(
			"searching for `%s`: reading http response body: %w",
			query,
			err,
		)
		return
	}

	if rsp.StatusCode != http.StatusOK {
		err = fmt.Errorf(
			"searching for `%s`: non-%d response status:\n%s",
			query,
			http.StatusOK,
			data,
		)
		return
	}

	if err = json.Unmarshal(data, &results); err != nil {
		err = fmt.Errorf(
			"searching for `%s`: unmarshaling http response: %w",
			query,
			err,
		)
		return
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Seeders > results[j].Seeders
	})
	return
}

type SearchResult struct {
	Name     string    `json:"name"`
	InfoHash InfoHash  `json:"infoHash"`
	Size     int       `json:"size"`
	Files    int       `json:"files"`
	Seeders  int       `json:"seeders"`
	Leechers int       `json:"leechers"`
	User     string    `json:"user"`
	Uploaded time.Time `json:"uploaded"`
}

// UnmarshalJSON implements the `json.Unmarshaler` interface.
func (result *SearchResult) UnmarshalJSON(data []byte) error {
	var payload struct {
		Name     string    `json:"name"`
		InfoHash InfoHash  `json:"info_hash"`
		Leechers intString `json:"leechers"`
		Seeders  intString `json:"seeders"`
		NumFiles intString `json:"num_files"`
		Size     intString `json:"size"`
		Username string    `json:"username"`
		Added    intString `json:"added"`
		Status   string    `json:"status"`
	}

	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	result.Name = payload.Name
	result.InfoHash = payload.InfoHash
	result.Leechers = int(payload.Leechers)
	result.Seeders = int(payload.Seeders)
	result.Files = int(payload.NumFiles)
	result.Size = int(payload.Size)
	result.User = payload.Username
	result.Uploaded = time.Unix(int64(payload.Added), 0).UTC()

	return nil
}

// intString simplifies JSON unmarshaling of integer fields that are represented
// as a string (e.g., `"123"` instead of `123`). apibay.org makes heavy use of
// these for integer fields like seeders, leechers, num_files, etc.
type intString int

// UnmarshalJSON implements the `json.Unmarshaler` interface.
func (i *intString) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*(*int)(i), err = strconv.Atoi(s)
	return
}
