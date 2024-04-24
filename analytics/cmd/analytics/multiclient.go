package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
)

type MultiClient struct {
	Cursor  int
	Lock    sync.RWMutex
	Clients []Client
}

func (c *MultiClient) Locate(
	ctx context.Context,
	addr string,
) (l Location, err error) {
	c.Lock.Lock()
	start := c.Cursor
	c.Cursor = (c.Cursor + 1) % len(c.Clients)
	c.Lock.Unlock()

	var errs []error
	for i := range c.Clients {
		if l, err = c.Clients[(i+start)%len(c.Clients)].Locate(
			ctx,
			addr,
		); err != nil {
			slog.Info(
				"multiclient: failed to locate addr",
				"err", err.Error(),
				"client", i,
				"addr", addr,
			)
			errs = append(errs, err)
			continue
		}
		return
	}
	err = fmt.Errorf(
		"locating addr `%s`: exhausted all `%d` clients: %w",
		addr,
		len(c.Clients),
		errors.Join(errs...),
	)
	return
}
