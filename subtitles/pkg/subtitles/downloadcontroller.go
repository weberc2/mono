package subtitles

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type DownloadController struct {
	Model        Model
	LoopInterval time.Duration
}

func (dc *DownloadController) Run(ctx context.Context) error {
	ticker := time.NewTicker(dc.LoopInterval)
	defer ticker.Stop()

	for {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("download controller: %w", err)
		}

		if err := dc.runLoop(ctx); err != nil {
			slog.Error("download controller", "err", err.Error())
		}

		<-ticker.C
	}
}

func (dc *DownloadController) runLoop(ctx context.Context) error {
	panic("control loop not configured")
}
