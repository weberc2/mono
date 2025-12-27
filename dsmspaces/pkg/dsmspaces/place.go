package dsmspaces

import "encoding/json"

type Place struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitzero"`
	Location    struct {
		Coordinates Coordinates `json:"coordinates"`
	} `json:"location"`
	Attributes map[string]float64 `json:"features"`
	Hours      Hours              `json:"hours"`
}

type Hours [TimeOfDayCount]bool

func (hours Hours) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Morning   bool `json:"MORNING"`
		Afternoon bool `json:"AFTERNOON"`
		Evening   bool `json:"EVENING"`
		LateNight bool `json:"LATENIGHT"`
	}{
		Morning:   hours[TimeOfDayMorning],
		Afternoon: hours[TimeOfDayAfternoon],
		Evening:   hours[TimeOfDayEvening],
		LateNight: hours[TimeOfDayLateNight],
	})
}

func (hours *Hours) UnmarshalJSON(data []byte) error {
	var tmp struct {
		Morning   bool `json:"MORNING"`
		Afternoon bool `json:"AFTERNOON"`
		Evening   bool `json:"EVENING"`
		LateNight bool `json:"LATENIGHT"`
	}
	err := json.Unmarshal(data, &tmp)
	hours[TimeOfDayMorning] = tmp.Morning
	hours[TimeOfDayAfternoon] = tmp.Afternoon
	hours[TimeOfDayEvening] = tmp.Evening
	hours[TimeOfDayLateNight] = tmp.LateNight
	return err
}

var placeSchema = must(compiler.Compile([]byte(placeSchemaDocument)))

const placeSchemaDocument = `{
  "$id": "place.json",
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "Place",
  "type": "object",
  "additionalProperties": false,
  "required": ["id", "name", "location", "features"],

  "properties": {
    "id": {
      "type": "string",
      "description": "Stable unique identifier"
    },

    "name": {
      "type": "string"
    },

    "location": {
      "type": "object",
      "additionalProperties": false,
      "required": ["coordinates"],

      "properties": {
        "coordinates": {
          "type": "object",
          "required": ["latitude", "longitude"],
          "properties": {
            "latitude": { "type": "number" },
            "longitude": { "type": "number" }
          }
        },

        "neighborhood": {
          "type": "string"
        },

        "address": {
          "type": "string"
        }
      }
    },

    "features": {
      "description": "Normalized features of the place, mapped to the same axes as intents",
      "type": "object",
      "additionalProperties": false,

      "properties": {
        "quiet": {
          "type": "number",
          "minimum": -1.0,
          "maximum": 1.0,
          "description": "-1 = very loud, +1 = very quiet"
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
          "maximum": 1.0,
          "description": "-1 = no coffee, +1 = core offering"
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

    "hours": {
      "type": "object",
      "additionalProperties": false,
      "properties": {
        "MORNING": { "type": "boolean" },
        "AFTERNOON": { "type": "boolean" },
        "EVENING": { "type": "boolean" },
        "LATENIGHT": { "type": "boolean" }
      }
    }
  }
}`
