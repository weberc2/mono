package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type API struct {
	Logger     *slog.Logger
	Client     *http.Client
	Transforms TransformStore
	Downloads  DownloadStore
}

func (api *API) Run(ctx Context, addr string) error {
	api.Logger.Info("starting api", "addr", addr)
	var mux http.ServeMux
	api.Bind(&mux)
	server := http.Server{Addr: addr, Handler: &mux}
	go func() {
		<-ctx.Done()

		// create a new, non-canceled context with a 5 second timeout to
		// gracefully shut down the http server.
		ctx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			api.Logger.Error("shutting down http server", "err", err.Error())
			if err := server.Close(); err != nil {
				api.Logger.Error(
					"force-closing http server",
					"err", err.Error(),
				)
			}
		}
	}()

	if err := server.ListenAndServe(); err != nil && !errors.Is(
		err,
		http.ErrServerClosed,
	) {
		return fmt.Errorf("running api server: %w", err)
	}
	return nil
}

func (api *API) Bind(mux *http.ServeMux) {
	mux.HandleFunc(
		"POST /transforms",
		api.handle("creating transform", api.createTransform),
	)
	mux.HandleFunc(
		"GET /search",
		api.handle("searching", api.search),
	)
	mux.HandleFunc(
		"GET /downloads/{download}",
		api.handle("fetching download", api.fetchDownload),
	)
	mux.HandleFunc(
		"POST /downloads",
		api.handle("creating download", api.createDownload),
	)
	mux.HandleFunc(
		"GET /downloads",
		api.handle("listing downloads", api.listDownloads),
	)
}

func (api *API) handle(
	message string,
	f func(*http.Request, *slog.Logger) response,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := api.Logger.WithGroup("request")
		logger.With(
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("query", r.URL.RawQuery),
		)

		rsp := f(r, logger)
		data, err := json.Marshal(rsp.body)
		if err != nil {
			logger.Error(
				"%s: marshaling response body",
				message,
				slog.String("err", err.Error()),
				slog.String("body", fmt.Sprintf("%#v", rsp.body)),
			)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Internal Server Error"}`))
			return
		}
		if rsp.status < http.StatusInternalServerError {
			logger.Info(message)
		} else {
			logger.Error(message)
		}
		w.WriteHeader(rsp.status)
		w.Write(data)
	}
}

func (api *API) listDownloads(
	r *http.Request,
	logger *slog.Logger,
) (rsp response) {
	downloads, err := api.Downloads.ListDownloads(r.Context())
	if err != nil {
		handleErr(logger, &rsp, err)
		return
	}

	rsp.status = http.StatusOK
	rsp.body = downloads
	return
}

func (api *API) fetchDownload(
	r *http.Request,
	logger *slog.Logger,
) (rsp response) {
	id := r.PathValue("download")
	logger.With(slog.String("infoHash", id))
	download, err := api.Downloads.FetchDownload(r.Context(), DownloadID(id))
	if err != nil {
		handleErr(logger, &rsp, err)
		return
	}

	rsp.status = http.StatusOK
	rsp.body = download
	return
}

func (api *API) createDownload(
	r *http.Request,
	logger *slog.Logger,
) (rsp response) {
	var err error
	var data []byte
	const bodyLimit = 1024 * 1024 * 1024
	if data, err = io.ReadAll(io.LimitReader(r.Body, bodyLimit)); err != nil {
		handleErr(logger, &rsp, fmt.Errorf("reading request body: %w", err))
		return
	}

	var d Download
	if err = json.Unmarshal(data, &d); err != nil {
		handleErr(
			logger,
			&rsp,
			fmt.Errorf("unmarshaling request body as download: %w", err),
		)
		return
	}

	if d, err = api.Downloads.CreateDownload(r.Context(), &d); err != nil {
		handleErr(logger, &rsp, err)
		return
	}

	if data, err = json.Marshal(&d); err != nil {
		handleErr(logger, &rsp, fmt.Errorf("marshaling download: %w", err))
		return
	}

	rsp.status = http.StatusCreated
	rsp.body = &d
	return
}

func (api *API) search(r *http.Request, logger *slog.Logger) (rsp response) {
	if query := strings.TrimSpace(r.URL.Query().Get("query")); query != "" {
		results, err := Search(api.Client, r.Context(), query)
		if err != nil {
			handleErr(logger, &rsp, err)
			return
		}

		rsp.status = http.StatusOK
		rsp.body = results
		return
	}

	rsp.status = http.StatusBadRequest
	rsp.body = &struct {
		Error   string `json:"error"`
		Details string `json:"details"`
	}{
		Error:   "Missing Required Parameter",
		Details: "Parameter `query` in query string is empty or missing",
	}
	return
}

func (api *API) createTransform(
	r *http.Request,
	logger *slog.Logger,
) (rsp response) {
	var err error
	var data []byte
	const bodyLimit = 1024 * 1024 * 1024
	if data, err = io.ReadAll(io.LimitReader(r.Body, bodyLimit)); err != nil {
		handleErr(logger, &rsp, fmt.Errorf("reading request body: %w", err))
		return
	}

	var transform Transform
	if err = json.Unmarshal(data, &transform); err != nil {
		handleErr(
			logger,
			&rsp,
			fmt.Errorf("unmarshaling request body: %w", err),
		)
		return
	}

	if transform, err = api.Transforms.CreateTransform(
		r.Context(),
		&transform,
	); err != nil {
		handleErr(logger, &rsp, err)
		return
	}

	if data, err = json.Marshal(&transform); err != nil {
		handleErr(logger, &rsp, err)
		return
	}

	rsp.status = http.StatusCreated
	rsp.body = json.RawMessage(data)
	return
}

type response struct {
	status int
	body   interface{}
}

func handleErr(logger *slog.Logger, rsp *response, err error) {
	logger.With(slog.String("err", err.Error()))
	rsp.status = http.StatusInternalServerError
	rsp.body = struct {
		Error string `json:"error"`
	}{
		Error: "Internal Server Error",
	}
}
