package main

import (
	"context"
	"net/http"
)

type Client struct {
	HTTP    *http.Client
	APIKey  string
	Locator LocatorFunc
}

func (c *Client) Locate(
	ctx context.Context,
	addr string,
) (l Location, err error) {
	return c.Locator(ctx, c.HTTP, addr, c.APIKey)
}

type LocatorFunc func(
	ctx context.Context,
	c *http.Client,
	addr string,
	apiKey string,
) (l Location, err error)

type LocationSource string

const (
	LocationSourceStack       LocationSource = "ipstack.com"
	LocationSourceGeolocation LocationSource = "ipgeolocation.io"
)

type Location struct {
	Source        LocationSource `json:"location_source"`
	ContinentCode string         `json:"continent_code"`
	ContinentName string         `json:"continent_name"`
	CountryCode   string         `json:"country_code"`
	CountryName   string         `json:"country_name"`
	RegionCode    string         `json:"region_code"`
	RegionName    string         `json:"region_name"`
	City          string         `json:"city"`
	Zip           string         `json:"zip"`
	Latitude      float64        `json:"latitude"`
	Longitude     float64        `json:"longitude"`
}

func (l *Location) updateMap(data map[string]interface{}) {
	data["continent_code"] = l.ContinentCode
	data["continent_name"] = l.ContinentName
	data["country_code"] = l.CountryCode
	data["country_name"] = l.CountryName
	data["region_code"] = l.RegionCode
	data["region_name"] = l.RegionName
	data["city"] = l.City
	data["zip"] = l.Zip
	data["latitude"] = l.Latitude
	data["longitude"] = l.Longitude
}
