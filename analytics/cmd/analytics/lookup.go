package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func lookup(
	ctx context.Context,
	c *http.Client,
	url string,
	v any,
) (err error) {
	var (
		r   *http.Request
		rsp *http.Response
	)
	if r, err = http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		url,
		nil,
	); err != nil {
		return err
	}
	defer func() { err = errors.Join(err, r.Body.Close()) }()

	if rsp, err = c.Do(r); err != nil {
		return err
	}

	var data []byte
	if data, err = io.ReadAll(rsp.Body); err != nil {
		err = fmt.Errorf("reading response body: %w", err)
		return
	}

	if rsp.StatusCode != http.StatusOK {
		err = fmt.Errorf("unsupported status code: %d\n%s", rsp.StatusCode, data)
		return
	}

	return json.Unmarshal(data, v)
}
