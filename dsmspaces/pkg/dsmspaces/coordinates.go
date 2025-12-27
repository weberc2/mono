package dsmspaces

import "math"

type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func (c *Coordinates) WithinRadius(other *Coordinates, radiusKM float64) bool {
	return c.DistanceKM(other) <= radiusKM
}

// DistanceKM computes the distance between two coordinates using the
// Haversine formula, returning the distance in kilometers.
func (c *Coordinates) DistanceKM(other *Coordinates) float64 {
	delta := Coordinates{
		Latitude:  rad(other.Latitude-c.Latitude) / 2,
		Longitude: rad(other.Longitude-c.Longitude) / 2,
	}
	a := math.Pow(math.Sin(delta.Latitude), 2) +
		math.Cos(rad(c.Latitude))*math.Cos(rad(other.Latitude))*
			math.Pow(math.Sin(delta.Longitude), 2)
	return earthRadiusKM * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

func rad(degrees float64) float64 {
	return degrees * math.Pi / 180
}

const earthRadiusKM = 6371.0
