package agent

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"golang.org/x/sync/errgroup"
)

type downloadParams struct {
	url  string
	file string
}

func addAptRepo(ctx context.Context, gpg, list downloadParams) error {
	group, ctx := errgroup.WithContext(ctx)
	group.Go(func() error { return fetchFile(ctx, gpg) })
	group.Go(func() error { return fetchFile(ctx, list) })
	if err := group.Wait(); err != nil {
		return fmt.Errorf("adding apt repo: %w", err)
	}

	if err := runCmd("[apt update]", "apt", "update"); err != nil {
		return fmt.Errorf("adding apt repo: %w", err)
	}

	return nil
}

func fetchFile(ctx context.Context, params downloadParams) error {
	req, err := http.NewRequestWithContext(ctx, "GET", params.url, nil)
	if err != nil {
		return fmt.Errorf("fetching file `%s`: %w", params.file, err)
	}
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("fetching file `%s`: %w", params.file, err)
	}
	defer close("response body", rsp.Body)

	file, err := os.Create(params.file)
	if err != nil {
		return fmt.Errorf(
			"fetching file `%s`: creating file: %v",
			params.file,
			err,
		)
	}
	defer close(params.file, file)

	if _, err := io.Copy(file, rsp.Body); err != nil {
		return fmt.Errorf(
			"fetching file `%s`: copying HTTP response body into file: %w",
			params.file,
			err,
		)
	}

	return nil
}

func close(name string, c io.Closer) {
	if err := c.Close(); err != nil {
		log.Printf("ERROR closing %s: %v", name, err)
	}
}
