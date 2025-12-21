package dsmspaces

import "fmt"

type PlaceType int

func (pt PlaceType) String() string {
	return placeTypeStrings[pt]
}

func (pt PlaceType) MarshalText() ([]byte, error) {
	return []byte(pt.String()), nil
}

func (pt *PlaceType) UnmarshalText(data []byte) error {
	for *(*int)(pt) = range placeTypeStrings {
		if placeTypeStrings[*pt] == string(data) {
			break
		}
	}
	return fmt.Errorf("invalid place type: %s", data)
}

var (
	placeTypeStrings = [...]string{
		PlaceTypePublic:   "PUBLIC",
		PlaceTypeBusiness: "BUSINESS",
	}
)

const (
	PlaceTypePublic PlaceType = iota
	PlaceTypeBusiness
)
