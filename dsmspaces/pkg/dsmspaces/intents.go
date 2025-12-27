package dsmspaces

import (
	"fmt"
)

type Intents struct {
	Attributes map[string]float64 `json:"attributes,omitzero"`
	Open       Availability       `json:"open,omitzero"`
	Near       Location           `json:"near,omitzero"`
}

type Location struct {
	Coordinates Opt[Coordinates] `json:"coordinates,omitzero"`
}

type Availability struct {
	TimeOfDay []TimeOfDay `json:"timeOfDay,omitzero"`
}

type TimeOfDay int

func (t *TimeOfDay) UnmarshalText(data []byte) error {
	for *(*int)(t) = range timeOfDayStrings {
		if string(data) == timeOfDayStrings[*t] {
			return nil
		}
	}

	return fmt.Errorf("unsupported time of day: %s", data)
}

func (t TimeOfDay) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

func (t TimeOfDay) String() string { return timeOfDayStrings[t] }

const (
	TimeOfDayMorning TimeOfDay = iota
	TimeOfDayAfternoon
	TimeOfDayEvening
	TimeOfDayLateNight
	TimeOfDayCount
)

var (
	timeOfDayStrings = [...]string{
		TimeOfDayMorning:   "MORNING",
		TimeOfDayAfternoon: "AFTERNOON",
		TimeOfDayEvening:   "EVENING",
		TimeOfDayLateNight: "LATENIGHT",
	}
)

func ValidateIntents(data []byte) error {
	result := intentsSchema.ValidateJSON(data)
	if !result.Valid {
		return fmt.Errorf("invalid intents json: %w", result)
	}
	return nil
}

var (
	intentsSchema = must(compiler.Compile([]byte(intentsSchemaDocument)))
)

const intentsSchemaDocument = `{
	"$id": "intents.json",
	"$schema": "https://json-schema.org/draft/2020-12/schema",
	"title": "Intents",
	"description": "Represents the desired filters to be applied when searching a collection of places",
	"type": "object",
	"additionalProperties": false,
	"required": [],
	"properties": {
		"attributes": {
			"description": "The attributes of a given place. The field's number reflects the intensity of the desire for a match. If a field is missing or zero, it indicates that the user doesn't care about that attribute.",
			"type": "object",
			"additionalProperties": false,
			"properties": {
				"quiet": {
					"type": "number",
					"minimum": -1.0,
					"maximum": 1.0
				},
				"kidFriendly": {
					"type": "number",
					"minimum": -1.0,
					"maximum": 1.0
				},
				"dogFriendly": {
					"type": "number",
					"minimum": -1.0,
					"maximum": 1.0
				},
				"readingFriendly": {
					"type": "number",
					"minimum": -1.0,
					"maximum": 1.0
				},
				"coffee": {
					"type": "number",
					"minimum": -1.0,
					"maximum": 1.0
				},
				"alcohol": {
					"type": "number",
					"minimum": -1.0,
					"maximum": 1.0
				},
				"screens": {
					"type": "number",
					"minimum": -1.0,
					"maximum": 1.0
				}
			}
		},
		"open": {
			"description": "A filter for the hours a place is open",
			"type": "object",
			"additionalProperties": false,
			"properties": {
				"timeOfDay": {
					"type": "array",
					"items": {
						"enum": [
							"MORNING",
							"AFTERNOON",
							"EVENING",
							"LATENIGHT"
						]
					},
					"minItems": 1
				}
			}
		},
		"near": {
			"description": "A filter for proximity to a location",
			"type": "object",
			"additionalProperties": false,
			"properties": {
				"coordinates": {
					"type": "object",
					"additionalProperties": false,
					"properties": {
						"latitude": {
							"type": "number",
							"minimum": -90,
							"maximum": 90
						},
						"longitude": {
							"type": "number",
							"minimum": -180,
							"maximum": 180
						}
					},
					"required": ["latitude", "longitude"]
				}
			}
		}
	}
}`
