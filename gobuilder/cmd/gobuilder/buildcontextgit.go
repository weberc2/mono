package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type BuildContextGit struct {
	Repository string `json:"repository"`
	Branch     string `json:"branch,omitempty"`
	Directory  string `json:"directory,omitempty"`
}

func (bcg *BuildContextGit) CopyToBuildDirectory(
	ctx context.Context,
	secrets *Secrets,
	directory string,
) error {
	var tmp string
	if bcg.Directory != "" {
		var err error
		if tmp, err = os.MkdirTemp("/tmp", "git-repo-*"); err != nil {
			return fmt.Errorf(
				"cloning git build context to build directory `%s`: "+
					"creating temporary repository directory: %w",
				directory,
				err,
			)
		}
		defer func() {
			if err := os.RemoveAll(tmp); err != nil {
				slog.Error(
					"cloning git build context to build directory; "+
						"cleaning up temporary repository directory",
					"err", err.Error(),
					"directory", tmp,
				)
			}
		}()
	} else {
		tmp = directory
	}

	// go-git doesn't support partial clones (and apparently Git may not
	// either?
	// https://github.com/go-git/go-git/issues/713#issuecomment-1616169675) so
	// instead we are cloning the full repository to the directory and then if
	// the build context is a subdirectory, then we'll set the build directory
	// to that subdirectory (by renaming the subdirectory with the name of the
	// build directory)

	// 1. clone
	if _, err := git.PlainCloneContext(
		ctx,
		tmp,
		false,
		&git.CloneOptions{
			URL:           bcg.Repository,
			Auth:          secrets.RepositoryAuthMethods[bcg.Repository],
			SingleBranch:  true,
			ReferenceName: plumbing.NewBranchReferenceName(bcg.Branch),
			Tags:          git.NoTags,
		},
	); err != nil {
		return fmt.Errorf(
			"cloning git build context to build directory `%s`: %w",
			tmp,
			err,
		)
	}

	// 2. ensure the build context directory is the build directory
	if bcg.Directory != "" {
		src := filepath.Join(tmp, bcg.Directory)
		if err := os.Rename(src, directory); err != nil {
			return fmt.Errorf(
				"cloning git build context to build directory: "+
					"moving `%s` to `%s`: %w",
				src,
				directory,
				err,
			)
		}
		slog.Info("moved directory", "src", src, "dst", directory)
	}

	return nil
}
