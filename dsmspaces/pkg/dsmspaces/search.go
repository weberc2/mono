package dsmspaces

import (
	"cmp"
	"slices"
)

func Search(places []Place, intents *Intents) (results []ScoredPlace) {
	if len(places) < 1 {
		return
	}

	const threshold = 0.0
	results = make([]ScoredPlace, len(places))
	for i := range places {
		results[i] = Score(&places[i], intents)
	}

	// sort the results by score in descending order (highest score first)
	slices.SortFunc(results, func(a, b ScoredPlace) int {
		return -cmp.Compare(a.Score, b.Score)
	})

	// truncate the results as soon as one of the scores drop below the
	// threshold
	for i := range results {
		// curve the scores--the first score is the highest as a consequence of
		// the previous sort, so we can divide everything by it as the curving
		// mechanism
		results[i].Score = results[i].Score / results[0].Score
		if results[i].Score < threshold {
			results = results[:i]
			return
		}
	}
	return
}

type ScoredPlace struct {
	Place           *Place             `json:"place"`
	Score           float64            `json:"score"`
	ExcludedReason  string             `json:"excludedReason,omitzero"`
	AttributeScores map[string]float64 `json:"attributeScores,omitzero"`
}

func Score(place *Place, intents *Intents) (result ScoredPlace) {
	result.Place = place

	// if the place is not open during the required times, exclude it by
	// returning 0
	for _, requiredTime := range intents.Open.TimeOfDay {
		if !place.Hours[requiredTime] {
			result.ExcludedReason = "not open during required time"
			return
		}
	}

	// if the intents specifies proximity to some coordinates, but the place is
	// not within a 1km radius to the coordinates, then exclude it by returning
	// 0
	if intendedCoords, ok := intents.Near.Coordinates.Get(); ok {
		if !place.Location.Coordinates.WithinRadius(&intendedCoords, 1) {
			result.ExcludedReason = "not near intended coords"
			return
		}
	}

	result.AttributeScores = make(map[string]float64, len(intents.Attributes))
	for attr, weight := range intents.Attributes {
		result.AttributeScores[attr] = place.Attributes[attr] * weight
		result.Score += place.Attributes[attr] * weight
	}
	return
}
