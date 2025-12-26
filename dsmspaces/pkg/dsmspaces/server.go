package dsmspaces

import (
	"dsmspaces/pkg/logger"
	"encoding/json"
	"net/http"
	"os"
)

type Server struct {
	Places       []Place
	IndexFile    string
	IntentParser IntentsParser
}

func NewServer(
	indexFile string,
	places []Place,
	openaiAPIKey string,
) (server Server) {
	server.IndexFile = indexFile
	server.Places = places
	server.IntentParser = NewIntentsParser(openaiAPIKey)
	return
}

func (s *Server) ServeIndex(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile(s.IndexFile)
	if err != nil {
		logger.Get(r.Context()).Error(
			"serving index",
			"action", "opening index file",
			"err", err.Error(),
		)
		httperr(w, http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "text/html")
	if _, err = w.Write(data); err != nil {
		logger.Get(r.Context()).Warn(
			"serving index",
			"action", "writing response body",
			"err", err.Error(),
		)
		return
	}
}

func (s *Server) Search(w http.ResponseWriter, r *http.Request) {
	var (
		l       = logger.Get(r.Context())
		intents Intents
		data    json.RawMessage
		err     error
	)

	if query := r.URL.Query().Get("q"); query != "" {
		l = l.With("query", query)
		if data, err = s.IntentParser.ParseIntentsJSON(
			r.Context(),
			query,
		); err != nil {
			l.Warn(
				"searching",
				"action", "parsing intents",
				"err", err.Error(),
				"intents", data,
			)
			httperr(w, http.StatusInternalServerError)
			return
		}

		l = l.With("intents", data)

		if err = json.Unmarshal(data, &intents); err != nil {
			l.Warn(
				"searching",
				"action", "unmarshaling intents",
				"err", err.Error(),
				"intents", string(data),
			)
			httperr(w, http.StatusInternalServerError)
			return
		}
	}

	results := Search(s.Places, &intents)
	if results == nil { // fix broken json marshaling of nil slices
		results = []ScoredPlace{}
	}
	if data, err = json.Marshal(results); err != nil {
		l.Warn(
			"searching",
			"action", "marshaling results",
			"err", err.Error(),
			"results", results,
		)
		httperr(w, http.StatusInternalServerError)
		return
	}

	l.Debug("searched places", "results", data)
	w.Header().Add("Content-Type", "application/json")
	w.Write(data)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger.Get(r.Context()).Info(
		"handling request",
		"method", r.Method,
		"path", r.URL.Path,
	)
	switch r.URL.Path {
	case "/", "/index.html":
		if r.Method == http.MethodGet {
			s.ServeIndex(w, r)
		} else {
			httperr(w, http.StatusMethodNotAllowed)
		}
	case "/search":
		if r.Method == http.MethodGet {
			s.Search(w, r)
		} else {
			httperr(w, http.StatusMethodNotAllowed)
		}
	default:
		httperr(w, http.StatusNotFound)
	}
}

func httperr(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}
