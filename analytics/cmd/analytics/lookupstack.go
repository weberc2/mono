package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
)

func lookupStack(
	ctx context.Context,
	c *http.Client,
	addr string,
	apiKey string,
) (l Location, err error) {
	var rsp struct {
		ContinentCode string  `json:"continent_code"`
		ContinentName string  `json:"continent_name"`
		CountryCode   string  `json:"country_code"`
		CountryName   string  `json:"country_name"`
		RegionCode    string  `json:"region_code"`
		RegionName    string  `json:"region_name"`
		City          string  `json:"city"`
		Zip           int     `json:"zip"`
		Latitude      float64 `json:"latitude"`
		Longitude     float64 `json:"longitude"`
	}

	url := fmt.Sprintf(
		"https://api.ipstack.com/%s?access_key=%s",
		addr,
		apiKey,
	)
	slog.Debug("locating addr with ipstack.com", "addr", addr, "url", url)
	if err = lookup(
		ctx,
		c,
		url,
		&rsp,
	); err != nil {
		err = fmt.Errorf("fetching ip address `%s` from ipstack: %w", addr, err)
	}

	l.Source = LocatorTypeStack
	l.ContinentCode = rsp.ContinentCode
	l.ContinentName = rsp.ContinentName
	l.CountryCode = rsp.CountryCode
	l.CountryName = rsp.CountryName
	l.RegionCode = rsp.RegionCode
	l.RegionName = rsp.RegionName
	l.City = rsp.City
	l.Zip = strconv.Itoa(rsp.Zip)
	l.Latitude = rsp.Latitude
	l.Longitude = rsp.Longitude
	return
}
