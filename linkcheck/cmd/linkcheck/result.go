package main

type Result struct {
	BaseURL       string `json:"baseURL"`
	TargetText    string `json:"targetText"`
	TargetURL     string `json:"targetURL"`
	URLParseError error  `json:"urlParseError,omitempty"`
	NetworkError  error  `json:"networkError,omitempty"`
	StatusCode    int    `json:"statusCode,omitempty"`
}
