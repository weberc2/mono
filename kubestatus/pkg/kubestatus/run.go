package kubestatus

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
)

//go:embed static
var staticFS embed.FS

func Run(ctx context.Context) error {
	var (
		level       slog.Level
		levelString = os.Getenv("LOG_LEVEL")
		addr        = os.Getenv("ADDR")
		ready       bool
		podClient   podCache
	)

	// Setup signal handling to cancel context on SIGINT (Ctrl+C) or SIGTERM
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		sig := <-sigChan
		slog.Info("received signal", "signal", sig)
		cancel()
	}()

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		clientset, err := newClientset()
		if err == nil {
			podClient = newPodCache(clientset)
			// Start the pod informer and wait for the initial cache sync before
			// marking the application as ready. This ensures ListPods() returns
			// cached pod entries instead of an empty list.
			podClient.Start(ctx)

			ready = true
		}
		return err
	})

	if levelString != "" {
		if err := level.UnmarshalText([]byte(levelString)); err != nil {
			return fmt.Errorf(
				"unmarshaling log level `%s`: %w",
				levelString,
				err,
			)
		}
	}

	if addr == "" {
		addr = ":8080"
	}

	// Register HTTP handlers (do this synchronously so handlers are ready
	// before the server starts).
	http.HandleFunc(
		"/health",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		},
	)

	http.HandleFunc(
		"/ready",
		func(w http.ResponseWriter, r *http.Request) {
			var (
				status = http.StatusOK
				body   = readyBody
			)
			if !ready {
				body = notReadyBody
				status = http.StatusServiceUnavailable
				w.WriteHeader(status)
				w.Write(body)
				return
			}
			w.WriteHeader(status)
			w.Write(body)
		},
	)

	// Serve static files (CSS, JS, etc.)
	staticSub, err := fs.Sub(staticFS, "static")
	if err != nil {
		return fmt.Errorf("failed to create static file system: %w", err)
	}
	http.Handle(
		"/static/",
		http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))),
	)

	// Serve index.html on /
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var (
			level  = slog.LevelInfo
			status = http.StatusOK
			attrs  []slog.Attr
			data   []byte
			err    error
		)

		// Accept both `/` and `/index.html` for the index page.
		if r.URL.Path != "/" && r.URL.Path != "/index.html" {
			// Not found is client error, log at Info level
			level = slog.LevelInfo
			status = http.StatusNotFound
			data = notFoundBody
			goto RETURN
		}

		if data, err = fs.ReadFile(staticFS, "static/index.html"); err != nil {
			level = slog.LevelError
			status = http.StatusInternalServerError
			data = internalServerErrorBody
			attrs = append(attrs, slog.String("err", err.Error()))
			goto RETURN
		}

	RETURN:
		slog.LogAttrs(r.Context(), level, "serving index", attrs...)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(status)
		w.Write(data)
	})

	http.HandleFunc("/pods/", func(w http.ResponseWriter, r *http.Request) {
		var (
			level  = slog.LevelInfo
			status = http.StatusOK
			attrs  []slog.Attr
			data   []byte
			pods   []*v1.Pod
			err    error
		)

		if !ready {
			attrs = append(attrs, slog.String("err", "service unavailable"))
			status = http.StatusServiceUnavailable
			data = notReadyBody
			goto RETURN
		}

		if pods, err = podClient.ListPods(); err == nil {
			slog.Debug("listing pods", "count", len(pods))
			type pod struct {
				Namespace string `json:"namespace"`
				Name      string `json:"name"`
			}
			podsByNode := make(map[string][]pod)
			for _, p := range pods {
				nodePods := podsByNode[p.Spec.NodeName]
				podsByNode[p.Spec.NodeName] = append(
					nodePods,
					pod{Namespace: p.Namespace, Name: p.Name},
				)
			}

			if data, err = json.Marshal(podsByNode); err == nil {
				goto RETURN
			}
		}

		level = slog.LevelError
		status = http.StatusInternalServerError
		data = internalServerErrorBody
		attrs = append(attrs, slog.String("err", err.Error()))

	RETURN:
		slog.LogAttrs(r.Context(), level, "listing pods", attrs...)
		w.WriteHeader(status)
		w.Write(data)
	})

	srv := &http.Server{
		Addr:         addr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	// Start server in its own goroutine so we can also listen for ctx.Done()
	g.Go(func() error {
		slog.Info("starting http server", "addr", addr)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			return err
		}
		return nil
	})

	// Shutdown the server when the context is canceled.
	g.Go(func() error {
		<-ctx.Done()
		slog.Info("context canceled, shutting down http server")
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	})

	return g.Wait()
}

var (
	notFoundBody            = []byte("Not Found")
	notReadyBody            = []byte("Not Ready")
	readyBody               = []byte("Ready")
	internalServerErrorBody = []byte("Internal Server Error")
)
