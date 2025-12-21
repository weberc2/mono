package dsmspaces

import (
	"cmp"
	"math"
	"slices"
)

func Search(
	places []Place,
	expectations map[Attr]float64,
	minScore float64,
) (results []ScoredPlace) {
	for i := range places {
		if score := Score(
			expectations,
			places[i].Attributes,
		); score > minScore {
			results = append(results, ScoredPlace{
				Place: &places[i],
				Score: score,
			})
		}
	}
	slices.SortFunc(results, func(a, b ScoredPlace) int {
		return -cmp.Compare(a.Score, b.Score)
	})
	return
}

func Score(expectations, attributes map[Attr]float64) (score float64) {
	// apply expectations by measuring how badly each is violated, letting the
	// worst violation decide
	score = 1.0
	for attr, expect := range expectations {
		actual := attributes[attr] // defaults to 0 if missing
		if s := Satisfaction(expect, actual); s < score {
			score = s
		}
	}
	return
}

func Satisfaction(expect, actual float64) float64 {
	if expect > 0 {
		return math.Min(actual/expect, 1.0)
	}
	if expect < 0 {
		return math.Min((1.0-actual)/math.Abs(expect), 1.0)
	}
	return 1.0
}
