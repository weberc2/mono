package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/weberc2/auth/pkg/auth"
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

func (c *Client) Refresh(refreshToken string) (*auth.TokenDetails, error) {
	data, err := json.Marshal(struct {
		RefreshToken string `json:"refreshToken"`
	}{refreshToken})
	if err != nil {
		return nil, fmt.Errorf("marshaling refresh token: %w", err)
	}
	rsp, err := c.HTTP.Post(
		c.BaseURL+"/api/refresh",
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
	var tokens auth.TokenDetails
	if err := json.Unmarshal(data, &tokens); err != nil {
		return nil, fmt.Errorf("invalid token refresh payload: %w", err)
	}
	return &tokens, nil
}

func (c *Client) Exchange(code string) (*auth.TokenDetails, error) {
	data, err := json.Marshal(&auth.Code{Code: code})
	if err != nil {
		return nil, fmt.Errorf("marshaling code: %w", err)
	}

	rsp, err := c.HTTP.Post(
		c.BaseURL+"/api/exchange",
		"application/json",
		bytes.NewReader(data),
	)
	if err != nil {
		return nil, fmt.Errorf("exchanging auth code: %w", err)
	}
	if rsp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"auth code exchange status: wanted `200`; found `%d`",
			rsp.StatusCode,
		)
	}
	defer rsp.Body.Close()

	data, err = ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading exchange response body: %w", err)
	}

	var tokens auth.TokenDetails
	if err := json.Unmarshal(data, &tokens); err != nil {
		return nil, fmt.Errorf("unmarshaling exchange response: %w", err)
	}

	return &tokens, nil
}
