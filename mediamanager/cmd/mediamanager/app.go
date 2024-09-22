package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type App struct {
	DownloadController DownloadController
	Server             http.Server
}

func (app *App) Run(ctx Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	serverErr := make(chan error)
	go func() {
		slog.Info("starting api server", "addr", app.Server.Addr)
		serverErr <- app.Server.ListenAndServe()
	}()

	slog.Info("starting download controller")
	for {
		select {
		case <-ctx.Done():
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			if err := app.Server.Shutdown(ctx); err != nil {
				return fmt.Errorf("running app: %w", err)
			}
			return nil
		case err := <-serverErr:
			return fmt.Errorf("running app: %w", err)
		case <-ticker.C:
			app.DownloadController.controlLoop(ctx)
		}
	}
}
