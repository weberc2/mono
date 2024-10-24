package mm

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type API struct {
	Addr      string
	Downloads DownloadStore
	Imports   ImportStore
	Logger    *slog.Logger
}

func (api *API) Run(ctx context.Context) error {
	var mux http.ServeMux
	mux.HandleFunc("GET /downloads", api.ListDownloads)
	mux.HandleFunc("POST /downloads/{infoHash}", api.CreateDownload)
	mux.HandleFunc("GET /imports", api.ListImports)
	mux.HandleFunc("POST /imports/{id}", api.CreateImport)

	server := http.Server{Addr: api.Addr, Handler: &mux}
	if server.Addr == "" {
		server.Addr = ":8080"
	}

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
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

	api.Logger.Info("starting downloads api", "addr", server.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(
		err,
		http.ErrServerClosed,
	) {
		return fmt.Errorf("running api server: %w", err)
	}

	return nil
}

func (api *API) CreateImport(w http.ResponseWriter, r *http.Request) {
	logger := api.requestLogger(r)
	defer func() {
		if err := r.Body.Close(); err != nil {
			logger.Error("closing request body: %w", err)
		}
	}()

	id := r.PathValue("id")
	logger = logger.With("import", id)

	data, err := io.ReadAll(io.LimitReader(r.Body, maxRequestBodySize))
	if err != nil {
		logger.Error("reading request body", "err", err.Error())
		http500(w)
		return
	}

	var imp Import
	if err := json.Unmarshal(data, &imp); err != nil {
		logger.Info(
			"unmarshaling request body",
			"err", err.Error(),
			"data", string(data),
		)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	imp.Files = nil
	if imp.Film != nil {
		imp.Files = append(
			imp.Files,
			ImportFile{
				Path:   imp.Film.PrimaryVideoFile,
				Status: ImportFileStatusPending,
			},
		)
		for i := range imp.Film.PrimarySubtitles {
			imp.Files = append(
				imp.Files,
				ImportFile{
					Path:   imp.Film.PrimarySubtitles[i].Path,
					Status: ImportFileStatusPending,
				},
			)
		}
	}
	imp.Status = ImportStatusPending
	imp.ID = ImportID(id)

	if err := api.Imports.CreateImport(r.Context(), &imp); err != nil {
		if As[*ImportExistsErr](err) != nil {
			logger.Info("creating import", "err", err.Error())
			http.Error(w, "Import Exists", http.StatusConflict)
		} else {
			logger.Error("creating import", "err", err.Error())
			http500(w)
		}
		return
	}

	logger.Info("created import", "import", imp.ID)

	if data, err = json.Marshal(&imp); err != nil {
		logger.Error("marshaling import", "err", err.Error())
		http500(w)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write(data); err != nil {
		logger.Error("writing response body", "err", err.Error())
	}
}

func (api *API) ListImports(w http.ResponseWriter, r *http.Request) {
	logger := api.requestLogger(r)
	defer func() {
		if err := r.Body.Close(); err != nil {
			logger.Error("closing request body: %w", err)
		}
	}()

	api.Logger.Debug("listing imports")
	imports, err := api.Imports.ListImports(r.Context())
	if err != nil {
		logger.Error("listing imports", "err", err.Error())
		http500(w)
		return
	}
	api.Logger.Debug("listed imports")

	data, err := json.Marshal(imports)
	if err != nil {
		logger.Error("marshaling imports", "err", err.Error())
		http500(w)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		logger.Error("writing response", "err", err.Error())
	}
	logger.Info("listed imports")
}

func (api *API) CreateDownload(w http.ResponseWriter, r *http.Request) {
	logger := api.requestLogger(r)
	defer func() {
		if err := r.Body.Close(); err != nil {
			logger.Error("closing request body", "err", err.Error())
		}
	}()

	// clear out any fields that ought not have been set
	download := Download{
		ID:     NewInfoHash(r.PathValue("infoHash")),
		Status: DownloadStatusPending,
	}

	logger = logger.With("infoHash", download.ID)

	if err := api.Downloads.CreateDownload(r.Context(), &download); err != nil {
		if AsDownloadExistsErr(err) != nil {
			logger.Info("creating download", "err", err.Error())
			http.Error(w, "Download Exists", http.StatusConflict)
		} else {
			logger.Error("creating download", "err", err.Error())
			http500(w)
		}
		return
	}
	logger.Info("created download", "infoHash", download.ID)

	data, err := json.Marshal(&download)
	if err != nil {
		logger.Error("marshaling download", "err", err.Error())
		http500(w)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Add("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		logger.Error("writing response", "err", err.Error())
	}
}

func (api *API) ListDownloads(w http.ResponseWriter, r *http.Request) {
	logger := api.requestLogger(r)
	defer func() {
		if err := r.Body.Close(); err != nil {
			logger.Error("closing request body: %w", err)
		}
	}()

	api.Logger.Debug("listing downloads")
	downloads, err := api.Downloads.ListDownloads(r.Context())
	if err != nil {
		logger.Error("listing downloads", "err", err.Error())
		http500(w)
		return
	}
	api.Logger.Debug("listed downloads")

	data, err := json.Marshal(downloads)
	if err != nil {
		logger.Error("marshaling downloads", "err", err.Error())
		http500(w)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		logger.Error("writing response", "err", err.Error())
	}
	logger.Info("listed downloads")
}

func (api *API) requestLogger(r *http.Request) *slog.Logger {
	requestID := r.Header.Get("X-Request-Id")
	if requestID == "" {
		var data [8]byte
		rand.Read(data[:])
		requestID = base64.RawStdEncoding.EncodeToString(data[:])
	}
	return api.Logger.With("request", requestID)
}

func http500(w http.ResponseWriter) {
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}

const maxRequestBodySize = 1024 * 1024 * 1024
