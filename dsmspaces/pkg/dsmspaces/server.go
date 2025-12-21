package dsmspaces

import (
	"dsmspaces/pkg/logger"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"slices"
)

type Server struct {
	Places       []Place
	Attributes   []byte
	IndexFile    string
	IntentParser IntentParser
}

func NewServer(indexFile string, places []Place, openaiAPIKey string) (server Server) {
	server.IndexFile = indexFile
	server.Places = places
	server.IntentParser = NewIntentParser(openaiAPIKey)

	attrs := make(map[Attr]struct{}, len(places))
	for i := range places {
		for attr := range places[i].Attributes {
			attrs[attr] = struct{}{}
		}
	}

	attributes := make([]Attr, 0, len(attrs))
	for attr := range attrs {
		attributes = append(attributes, attr)
	}
	slices.Sort(attributes)

	var err error
	if server.Attributes, err = json.Marshal(attributes); err != nil {
		panic(fmt.Sprintf("marshaling attributes: %v", err))
	}
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

func (s *Server) ListAttributes(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	if _, err := w.Write(s.Attributes); err != nil {
		logger.Get(r.Context()).Warn(
			"listing attributes",
			"action", "writing response body",
			"err", err.Error(),
		)
	} else {
		logger.Get(r.Context()).Info("listing attributes")
	}
}

func (s *Server) Search(w http.ResponseWriter, r *http.Request) {
	var (
		l            = logger.Get(r.Context())
		expectations map[Attr]float64
		data         []byte
		err          error
	)

	if query := r.URL.Query().Get("q"); query != "" {
		if expectations, err = s.IntentParser.ParseIntent(
			r.Context(),
			query,
		); err != nil {
			l.Warn("searching", "action", "parsing intent", "err", err.Error())
			httperr(w, http.StatusInternalServerError)
			return
		}
		/*	} else {
			defer func() {
				if err := r.Body.Close(); err != nil {
					l.Warn("searching", "action", "closing request body", "err", err.Error())
				}
			}()

			const limit = 1024 * 1024
			if data, err = io.ReadAll(io.LimitReader(r.Body, limit)); err != nil {
				l.Warn("searching", "action", "reading request body", "err", err.Error())
				httperr(w, http.StatusInternalServerError)
				return
			}

			if err := json.Unmarshal(data, &expectations); err != nil {
				l.Info("searching", "action", "unmarshaling request body", "err", err.Error())
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
		*/
	}
	results := Search(s.Places, expectations, 0.0)
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
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

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
	case "/attributes":
		if r.Method == http.MethodGet {
			s.ListAttributes(w, r)
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
