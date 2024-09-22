package main

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type URL url.URL

func (u *URL) MarshalJSON() ([]byte, error) {
	return json.Marshal((*url.URL)(u).String())
}

func (u *URL) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("unmarshaling URL: %w", err)
	}

	if err := (*url.URL)(u).UnmarshalBinary([]byte(s)); err != nil {
		return fmt.Errorf("unmarshaling URL: %w", err)
	}

	return nil
}

func (u *URL) String() string { return (*url.URL)(u).String() }
