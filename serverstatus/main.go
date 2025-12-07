package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	v1 "k8s.io/api/core/v1"
)

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	var (
		level       slog.Level
		levelString = os.Getenv("LOG_LEVEL")
		addr        = os.Getenv("ADDR")
	)
	if err := level.UnmarshalText([]byte(levelString)); err != nil {
		return fmt.Errorf("unmarshaling log level `%s`: %w", levelString, err)
	}

	if addr == "" {
		addr = ":8080"
	}

	clientset, err := newClientset()
	if err != nil {
		return err
	}
	podClient := newPodCache(clientset)

	http.HandleFunc("/pods/", func(w http.ResponseWriter, r *http.Request) {
		var (
			level  = slog.LevelInfo
			status = http.StatusOK
			attrs  []slog.Attr
			data   []byte
			pods   []*v1.Pod
			err    error
		)
		if pods, err = podClient.ListPods(); err == nil {
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
		slog.LogAttrs(ctx, level, "listing pods", attrs...)
		w.WriteHeader(status)
		w.Write(data)
	})

	srv := &http.Server{
		Addr:         addr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	slog.Info("starting http server", "addr", addr)
	return srv.ListenAndServe()
}

var internalServerErrorBody = []byte("Internal Server Error")
