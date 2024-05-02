package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
)

// MultiLocator wraps multiple `Locator`s in a single `Locator`-like interface.
// It attempts them in a round-robin fashion so as to consume quota somewhat
// evenly.
type MultiLocator struct {
	// cursor keeps track of which locator was most recently attempted.
	cursor int

	// lock locks `cursor` so only one goroutine reads/writes to it at a time.
	lock sync.Mutex

	// Locators is the collection of locators to use.
	Locators []Locator
}

// Locate fetches the location of an IP address from the supported locators in a
// round-robin manner. Every call to `Locate()` will begin attempting lookups
// with a different locator. It will only return an error if all locators failed
// to find a location for the provided address (in which case it will return an
// error which wraps all of the individual locators' errors).
func (l *MultiLocator) Locate(
	ctx context.Context,
	c *http.Client,
	addr string,
) (location Location, err error) {
	l.lock.Lock()
	start := l.cursor
	l.cursor = (l.cursor + 1) % len(l.Locators)
	l.lock.Unlock()

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
