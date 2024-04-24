package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
)

func lookupGeolocation(
	ctx context.Context,
	c *http.Client,
	addr string,
	apiKey string,
) (l Location, err error) {
	var rsp struct {
		ContinentCode string `json:"continent_code"`
		ContinentName string `json:"continent_name"`
		CountryCode   string `json:"country_code2"`
		CountryName   string `json:"country_name"`
		RegionCode    string `json:"state_prov"`
		RegionName    string `json:"region_name"`
		City          string `json:"city"`
		Zip           string `json:"zipcode"`
		Latitude      string `json:"latitude"`
		Longitude     string `json:"longitude"`
	}

	if err = lookup(
		ctx,
		c,
		fmt.Sprintf(
			"https://api.ipgeolocation.io/ipgeo?apiKey=%[2]s&ip=%[1]s",
			addr,
			apiKey,
		),
		&rsp,
	); err != nil {
		err = fmt.Errorf(
			"fetching location for ip address `%s` from ipstack.com: %w",
			addr,
			err,
		)
		return
	}

	l.ContinentCode = rsp.ContinentCode
	l.ContinentName = rsp.ContinentName
	l.CountryCode = rsp.CountryCode
	l.CountryName = rsp.CountryName
	l.RegionCode = rsp.RegionCode
	l.RegionName = rsp.RegionName
	l.City = rsp.City
	l.Zip = rsp.Zip
	if l.Latitude, err = strconv.ParseFloat(rsp.Latitude, 64); err != nil {
		err = fmt.Errorf("parsing latitude `%s`: %w", rsp.Latitude, err)
		return
	}
	if l.Longitude, err = strconv.ParseFloat(rsp.Longitude, 64); err != nil {
		err = fmt.Errorf("parsing longitude `%s`: %w", rsp.Longitude, err)
	}
	return

	// https://ipgeolocation.io/documentation/ip-geolocation-api.html
	//
	//	{
	//	    "ip": "8.8.8.8",
	//	    "hostname": "dns.google",
	//	    "continent_code": "NA",
	//	    "continent_name": "North America",
	//	    "country_code2": "US",
	//	    "country_code3": "USA",
	//	    "country_name": "United States",
	//	    "country_capital": "Washington, D.C.",
	//	    "state_prov": "California",
	//	    "district": "Santa Clara",
	//	    "city": "Mountain View",
	//	    "zipcode": "94043-1351",
	//	    "latitude": "37.42240",
	//	    "longitude": "-122.08421",
	//	    "is_eu": false,
	//	    "calling_code": "+1",
	//	    "country_tld": ".us",
	//	    "languages": "en-US,es-US,haw,fr",
	//	    "country_flag": "https://ipgeolocation.io/static/flags/us_64.png",
	//	    "geoname_id": "6301403",
	//	    "isp": "Google LLC",
	//	    "connection_type": "",
	//	    "organization": "Google LLC",
	//	    "asn": "AS15169",
	//	    "currency": {
	//	        "code": "USD",
	//	        "name": "US Dollar",
	//	        "symbol": "$"
	//	    },
	//	    "time_zone": {
	//	        "name": "America/Los_Angeles",
	//	        "offset": -8,
	//	        "current_time": "2020-12-17 07:49:45.872-0800",
	//	        "current_time_unix": 1608220185.872,
	//	        "is_dst": false,
	//	        "dst_savings": 1
	//	    }
	//	}
}
