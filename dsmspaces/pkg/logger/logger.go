package logger

import (
	"context"
	"log/slog"
)

func Set(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

func Get(ctx context.Context) (l *slog.Logger) {
	if v := ctx.Value(loggerKey); v != nil {
		if l = v.(*slog.Logger); l != nil {
			return
		}
	}
	l = slog.Default()
	return
}

type loggerKeyType string

const loggerKey loggerKeyType = "loggerKey"
