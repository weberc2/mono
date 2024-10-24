package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"mediamanager/pkg/mm"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
)

type API struct {
	Downloads mm.DownloadStore
	Imports   mm.ImportStore
	Logger    *slog.Logger
}

func (api *API) Run(ctx context.Context, addr string) error {
	server := http.Server{Addr: addr, Handler: api.Handler()}

	done := make(chan struct{})

	go func() {
		if e := server.ListenAndServe(); !errors.Is(e, http.ErrServerClosed) {
			api.Logger.Error("serving http", "err", e.Error())
		}
		api.Logger.Info("http server shutdown")
		done <- struct{}{}
	}()

	// block until the http server encounters an error or until the context is
	// canceled. on context cancellation, shutdown the http server (gracefully,
	// if possible).
	select {
	case <-ctx.Done():
		sdc, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := server.Shutdown(sdc); err != nil &&
			!errors.Is(err, context.Canceled) &&
			!errors.Is(err, context.DeadlineExceeded) {

			return fmt.Errorf("running api: shutting down http server: %w", err)
		}
	case <-done:
	}
	return nil
}

func (api *API) Handler() http.Handler {
	var mux http.ServeMux
	config := huma.DefaultConfig("mm", "v0.0.1-alpha0")
	registry := Registry{API: api, Huma: humago.New(&mux, config)}
	OperationImportList.Register(&registry)
	OperationImportCreate.Register(&registry)
	OperationImportDelete.Register(&registry)
	OperationDownloadList.Register(&registry)
	OperationDownloadCreate.Register(&registry)
	OperationDownloadDelete.Register(&registry)
	return &mux
}

type Registry struct {
	API  *API
	Huma huma.API
}

type Operation[I, O any] struct {
	Huma    huma.Operation
	Handler func(api *API, ctx context.Context, input *I) (*O, error)
}

func (op Operation[I, O]) Register(r *Registry) {
	huma.Register(r.Huma, op.Huma, func(ctx context.Context, i *I) (*O, error) {
		return op.Handler(r.API, ctx, i)
	})
}

type OperationID string
