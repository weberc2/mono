package main

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"time"
)

const (
	errorCodeOK                    = 0
	errorCodeInsufficientArguments = 1
	errorCodeInvalidURL            = 2
	errorCodeToplevelNetworkError  = 3
	errorCodeFailuresDetected      = 4
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	start := time.Now()
	slog.Debug("starting", "time", start)
	defer func() { slog.Debug("completed", "elapsed", time.Since(start)) }()

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "USAGE: linkcheck URL\n")
		os.Exit(errorCodeInsufficientArguments)
	}

	u, err := url.Parse(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "parsing base url: %v", err)
		os.Exit(errorCodeInvalidURL)
	}

	crawler := NewCrawler(u.Host)

	// Prepare the visitor.
	errorCounter := new(CountVisitor)
	if isTTY() {
		// If we're a TTY, print everything but also count errors
		crawler.SetVisitor(NewMultiVisitor(
			new(PrettyResultPrinter),
			new(ErrorVisitor).SetInner(errorCounter),
		))
	} else {
		// If we're not a TTY, then count errors and only print them
		crawler.SetVisitor(new(ErrorVisitor).SetInner(NewMultiVisitor(
			errorCounter,
			new(JSONResultPrinter),
		)))
	}

	if err := crawler.Crawl(u); err != nil {
		fmt.Fprintf(os.Stderr, "toplevel network error: %v", err)
		os.Exit(errorCodeToplevelNetworkError)
	}

	if errorCount := errorCounter.GetCount(); errorCount > 0 {
		os.Exit(errorCodeFailuresDetected)
	}

	os.Exit(errorCodeOK)
}
