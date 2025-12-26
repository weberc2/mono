package dsmspaces

import "math"

type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func (c *Coordinates) WithinRadius(other *Coordinates, radiusKM float64) bool {
	return c.HaversineKM(other) <= radiusKM
}

func (c *Coordinates) HaversineKM(other *Coordinates) float64 {
	delta := Coordinates{
		Latitude:  rad(other.Latitude-c.Latitude) / 2,
		Longitude: rad(other.Longitude-c.Longitude) / 2,
	}

	a := math.Sin(delta.Latitude)*math.Sin(delta.Longitude) +
		math.Cos(rad(c.Latitude))*
			math.Cos(rad(other.Latitude))*
			math.Sin(delta.Longitude)*
			math.Sin(delta.Longitude)

	return earthRadiusKM * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

func rad(degrees float64) float64 {
	return degrees * math.Pi / 180
}

const earthRadiusKM = 6371.0
