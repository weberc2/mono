package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
)

type MultiLocator struct {
	Cursor   int
	Lock     sync.RWMutex
	Locators []Locator
}

func (l *MultiLocator) Locate(
	ctx context.Context,
	c *http.Client,
	addr string,
) (location Location, err error) {
	l.Lock.Lock()
	start := l.Cursor
	l.Cursor = (l.Cursor + 1) % len(l.Locators)
	l.Lock.Unlock()

	var errs []error
	for i := range l.Locators {
		locator := &l.Locators[(i+start)%len(l.Locators)]
		if location, err = locator.Locate(ctx, c, addr); err != nil {
			slog.Info(
				"multilocator: failed to locate addr",
				"err", err.Error(),
				"addr", addr,
				"locatorUser", locator.User,
				"locatorIdentityProvider", locator.IdentityProvider,
			)
			errs = append(errs, err)
			continue
		}
		slog.Debug(
			"multi-client locate",
			"addr", addr,
			"locatorUser", locator.User,
			"locatorIdentityProvider", locator.IdentityProvider,
			"location", &location,
		)
		return
	}
	err = fmt.Errorf(
		"locating addr `%s`: exhausted all `%d` locators: %w",
		addr,
		len(l.Locators),
		errors.Join(errs...),
	)
	return
}
