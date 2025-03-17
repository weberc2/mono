package opensubtitles

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Client struct {
	HTTP      *http.Client
	APIKey    string
	UserAgent string
}

func (c *Client) Search(
	ctx context.Context,
	query *Query,
) (subtitles []Subtitle, err error) {
	var req *http.Request
	if req, err = http.NewRequestWithContext(
		ctx,
		"GET",
		"https://api.opensubtitles.com/api/v1/subtitles?"+query.String(),
		nil,
	); err != nil {
		err = fmt.Errorf("searching opensubtitles: preparing request: %w", err)
		return
	}

	req.Header.Add("User-Agent", c.UserAgent)
	req.Header.Add("Api-Key", c.APIKey)

	var rsp *http.Response
	if rsp, err = c.HTTP.Do(req); err != nil {
		err = fmt.Errorf("searching opensubtitles: %w", err)
		return
	}

	defer func() {
		if e := rsp.Body.Close(); e != nil {
			err = errors.Join(
				err,
				fmt.Errorf(
					"searching opensubtitles: closing response body: %w",
					e,
				),
			)
		}
	}()

	var data []byte
	if data, err = io.ReadAll(io.LimitReader(rsp.Body, 1024*1024)); err != nil {
		err = fmt.Errorf(
			"searching opensubtitles: reading response body: %w",
			err,
		)
		return
	}

	if rsp.StatusCode < 200 || rsp.StatusCode > 299 {
		err = fmt.Errorf(
			"searching opensubtitles: status code `%d`: %s",
			rsp.StatusCode,
			data,
		)
		return
	}

	var payload struct {
		Data []Subtitle `json:"data"`
	}

	if err = json.Unmarshal(data, &payload); err != nil {
		err = fmt.Errorf(
			"searching opensubtitles: unmarshaling subtitles: %w",
			err,
		)
		return
	}

	subtitles = payload.Data
	return
}

type Query struct {
	Type            QueryType
	Title           string
	Year            string
	Season          string
	Episode         string
	Languages       string
	HearingImpaired bool
}

func (q *Query) Values() (values url.Values) {
	values = url.Values{}
	if q.Type != "" {
		values.Add("type", string(q.Type))
	}
	if q.Title != "" {
		values.Add("query", q.Title)
	}
	if q.Year != "" {
		values.Add("year", q.Year)
	}
	if q.Season != "" {
		values.Add("season_number", q.Season)
	}
	if q.Episode != "" {
		values.Add("episode_number", q.Episode)
	}
	if q.Languages != "" {
		values.Add("languages", q.Languages)
	}
	if !q.HearingImpaired {
		values.Add("hearing_impaired", "exclude")
	}
	return
}

func (q *Query) String() string { return q.Values().Encode() }

type QueryType string

const (
	QueryTypeAll     QueryType = "all"
	QueryTypeEpisode QueryType = "episode"
	QueryTypeMovie   QueryType = "movie"
)

type Subtitle struct {
	ID           string   `json:"subtitle_id"`
	Language     string   `json:"language"`
	FrameRate    float64  `json:"fps"`
	Ratings      float64  `json:"ratings"`
	FromTrusted  bool     `json:"from_trusted"`
	AITranslated bool     `json:"ai_translated"`
	Slug         string   `json:"slug"`
	Release      string   `json:"release"`
	Uploader     Uploader `json:"uploader"`
}

func (s *Subtitle) UnmarshalJSON(data []byte) error {
	type subtitle Subtitle
	return json.Unmarshal(data, &struct {
		Attributes *subtitle `json:"attributes"`
	}{
		Attributes: (*subtitle)(s),
	})
}

type Uploader struct {
	ID   int    `json:"uploader_id"`
	Name string `json:"name"`
	Rank string `json:"rank"`
}
