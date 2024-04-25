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
	Clients []NamedClient
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
		client := &c.Clients[(i+start)%len(c.Clients)]
		if l, err = client.Locate(ctx, addr); err != nil {
			slog.Info(
				"multiclient: failed to locate addr",
				"err", err.Error(),
				"client", client.Name,
				"addr", addr,
			)
			errs = append(errs, err)
			continue
		}
		slog.Debug(
			"multi-client locate",
			"client", client.Name,
			"addr", addr,
			"location", &l,
		)
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

type NamedClient struct {
	Name string
	Client
}
