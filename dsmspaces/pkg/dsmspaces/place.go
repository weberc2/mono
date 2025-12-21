package dsmspaces

type Place struct {
	ID         string           `json:"id"`
	Name       string           `json:"name"`
	Type       PlaceType        `json:"placeType"`
	Attributes map[Attr]float64 `json:"attributes"`
}
