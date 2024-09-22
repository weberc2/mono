package main

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
)

type API struct {
	Downloads DownloadStore
}

func (api *API) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /downloads", api.CreateDownload)
}

func (api *API) CreateDownload(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := r.Body.Close(); err != nil {
			slog.Error(
				"creating download: closing request body",
				"err", err.Error(),
				"component", "API",
				"download", r.PathValue("download"),
			)
		}
	}()

	const maxBodySize = 1024 * 1024 * 1024 // 1GiB
	data, err := io.ReadAll(io.LimitReader(r.Body, maxBodySize))
	if err != nil {
		slog.Error(
			"creating download: reading request body",
			"err", err.Error(),
			"component", "API",
			"download", r.PathValue("download"),
		)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// TODO: validate the API spec. if we encounter a JSON unmarshaling error
	// at this point, it's a program error--not a user-input error--and we
	// should return 500 rather than 400.
	var spec DownloadSpec
	if err := json.Unmarshal(data, &spec); err != nil {
		slog.Error(
			"creating download: unmarshaling request body",
			"err", err.Error(),
			"component", "API",
			"download", r.PathValue("download"),
			"data", string(data),
		)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	download, err := api.Downloads.CreateDownload(
		r.Context(),
		&Download{
			ID:   DownloadID(spec.Source.InfoHash().String()),
			Spec: spec,
		},
	)
	if err != nil {
		slog.Error(
			"creating download: storing created download",
			"err", err.Error(),
			"component", "api",
			"download", r.PathValue("download"),
		)

		// if the error is NOT an HTTP error, then write it as an internal
		// server error.
		var httpErr HTTPError
		if !errors.As(err, &httpErr) {
			http.Error(
				w,
				"Internal Server Error",
				http.StatusInternalServerError,
			)
			return
		}

		// otherwise get the HTTP details and write them to the response writer
		details := httpErr.Details()
		data, err := json.Marshal(&details)
		if err != nil {
			slog.Error(
				"creating download: marshaling http error details",
				"err", err.Error(),
				"component", "api",
				"download", r.PathValue("download"),
				"httpErr", httpErr.Error(),
			)
			http.Error(
				w,
				"Internal Server Error",
				http.StatusInternalServerError,
			)
			return
		}
		w.WriteHeader(details.Status)
		w.Write(data)
		return
	}

	download.Status = DownloadStatusPending

	if data, err = json.Marshal(&download); err != nil {
		slog.Error(
			"creating download: marshaling download",
			"err", err.Error(),
			"component", "API",
			"download", download.ID,
		)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(data)
}

type HTTPError interface {
	error
	Details() HTTPErrorDetails
}

type HTTPErrorDetails struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}
