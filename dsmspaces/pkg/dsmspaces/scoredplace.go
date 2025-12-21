package dsmspaces

type ScoredPlace struct {
	Place *Place  `json:"place"`
	Score float64 `json:"score"`
}