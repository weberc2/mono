package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type Client struct {
	HTTP    http.Client
	BaseURL string
}

func DefaultClient(baseURL string) Client {
	return Client{
		HTTP:    http.Client{Timeout: 10 * time.Second},
		BaseURL: baseURL,
	}
}

type Tokens struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `jsno:"refreshToken"`
}

func (c *Client) Refresh(refreshToken string) (*Tokens, error) {
	data, err := json.Marshal(struct {
		RefreshToken string `json:"refreshToken"`
	}{refreshToken})
	if err != nil {
		return nil, fmt.Errorf("marshaling refresh token: %w", err)
	}
	rsp, err := c.HTTP.Post(
		c.BaseURL+"/refresh",
		"application/json",
		bytes.NewReader(data),
	)
	if err != nil {
		return nil, fmt.Errorf("refreshing access token: %w", err)
	}
	if rsp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"token refresh response status: wanted `200`; found `%d`",
			rsp.StatusCode,
		)
	}
	defer rsp.Body.Close()

	data, err = ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading token refresh response body: %w", err)
	}
	var tokens Tokens
	if err := json.Unmarshal(data, &tokens); err != nil {
		return nil, fmt.Errorf("invalid token refresh payload: %w", err)
	}
	return &tokens, nil
}
