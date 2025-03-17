package log

import (
	"context"
	"log/slog"

	"github.com/go-logr/logr"
)

func Context(ctx context.Context, logger *slog.Logger) context.Context {
	return logr.NewContextWithSlogLogger(ctx, logger)
}

func FromContext(ctx context.Context) (logger *slog.Logger) {
	return logr.FromContextAsSlogLogger(ctx)
}

type contextKeyType string

const contextKey contextKeyType = "LOGGER"
