package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Locator provides locations for IP addresses.
type Locator struct {
	// Type is the type of locator.
	Type LocatorType `json:"type"`

	// User is the user account associated with the locator. This is metadata
	// for identifying the locator in logging messages.
	User string `json:"user"`

	// IdentityProvider is the identity provider for the user. A given user
	// might have an account with a location provider under multiple identity
	// providers (e.g., a location provider might have separate accounts for
	// foo@example.com with identity providers Google and GitHub). This is
	// metadata for identifying the locator in logging messages.
	IdentityProvider string `json:"identityProvider"`

	// APIKey is the APIKey for the locator.
	APIKey string `json:"apiKey"`
}

// Locate locates an IP address using the geolocation service corresponding to
// `Locator.Type`.
func (l *Locator) Locate(
	ctx context.Context,
	c *http.Client,
	addr string,
) (location Location, err error) {
	if locate, ok := lookupFuncsByType[l.Type]; ok {
		return locate(ctx, c, addr, l.APIKey)
	}
	err = fmt.Errorf("invalid locator type: %s", l.Type)
	return
}

// LocatorType is the type of locator. Each type corresponds to an online IP
// geolocation service.
type LocatorType string

const (
	LocatorTypeStack       LocatorType = "ipstack.com"
	LocatorTypeGeolocation LocatorType = "ipgeolocation.io"
)

// UnmarshalJSON implements the `json.Unmarshaler` interface. In particular, it
// validates that the provided locator type is supported.
func (locatorType *LocatorType) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, (*string)(locatorType)); err != nil {
		return fmt.Errorf("unmarshaling locator type: %w", err)
	}
	if _, ok := lookupFuncsByType[*locatorType]; !ok {
		return fmt.Errorf(
			"unmarshaling locator type: %w",
			&InvalidLocatorTypeErr{Type: *locatorType},
		)
	}
	return nil
}

// InvalidLocatorTypeErr is returned when an unsupported `LocatorType` is
// passed.
type InvalidLocatorTypeErr struct {
	// Type is the unsupported locator type.
	Type LocatorType
}

// Error implements the `error` interface.
func (err *InvalidLocatorTypeErr) Error() string {
	return fmt.Sprintf("invalid locator type: %s", string(err.Type))
}

// Location contains the location information for an IP address.
type Location struct {
	Source        LocatorType `json:"location_source"`
	ContinentCode string      `json:"continent_code"`
	ContinentName string      `json:"continent_name"`
	CountryCode   string      `json:"country_code"`
	CountryName   string      `json:"country_name"`
	RegionCode    string      `json:"region_code"`
	RegionName    string      `json:"region_name"`
	City          string      `json:"city"`
	Zip           string      `json:"zip"`
	Latitude      float64     `json:"latitude"`
	Longitude     float64     `json:"longitude"`
}

var lookupFuncsByType = map[LocatorType]lookupFunc{
	LocatorTypeStack:       lookupStack,
	LocatorTypeGeolocation: lookupGeolocation,
}

type lookupFunc func(
	ctx context.Context,
	c *http.Client,
	addr string,
	apiKey string,
) (l Location, err error)
