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
	v interface{},
) (err error) {
	var (
		r    *http.Request
		rsp  *http.Response
		data []byte
	)
	if r, err = http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		url,
		nil,
	); err != nil {
		return
	}

	if rsp, err = c.Do(r); err != nil {
		return
	}
	defer func() { err = errors.Join(err, rsp.Body.Close()) }()

	if data, err = io.ReadAll(rsp.Body); err != nil {
		err = fmt.Errorf("reading response body: %w", err)
		return
	}

	if rsp.StatusCode != http.StatusOK {
		err = fmt.Errorf(
			"unexpected status code `%d`:\n%s",
			rsp.StatusCode,
			data,
		)
		return
	}

	if err = json.Unmarshal(data, v); err != nil {
		err = fmt.Errorf("unmarshaling response body: %w", err)
	}
	return
}
