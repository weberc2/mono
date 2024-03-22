package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

// EditSecret fetches the secret data for the secret identified by the namespace
// (if omitted, the current namespace will be attempted) and secret name, opens
// the data in an editor, and then requests the API server update the secret
// based on the changes.
func EditSecret(ctx context.Context, namespace, secretName string) error {
	clientset, defaultNamespace, err := loadClientset()
	if err != nil {
		return fmt.Errorf("editing secret `%s`: %w", secretName, err)
	}

	if namespace == "" {
		namespace = defaultNamespace
	}

	secrets := clientset.CoreV1().Secrets(namespace)

	original, err := secrets.Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("editing secret: %w", err)
	}

	if err := os.MkdirAll(tmpDir, 0700); err != nil {
		return fmt.Errorf(
			"editing secret `%s/%s`: creating tmp dir: %w",
			namespace,
			secretName,
			err,
		)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			slog.Error(
				"editing secret; cleaning up tmp dir",
				"err", err.Error(),
				"namespace", namespace,
				"secret", secretName,
			)
		}
	}()

	tmpFile, err := os.CreateTemp(tmpDir, secretName+"-*")
	if err != nil {
		return fmt.Errorf(
			"editing secret `%s/%s`: creating tmp file: %w",
			namespace,
			secretName,
			err,
		)
	}
	defer func() {
		if err := tmpFile.Close(); err != nil {
			slog.Error(
				"editing secret; closing tmp file",
				"err", err.Error(),
				"namespace", namespace,
				"secret", secretName,
			)
		}
	}()

	tmpFilePath := tmpFile.Name()

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	data, err := marshalSecret(original)
	if err != nil {
		return fmt.Errorf(
			"editing secret: marshaling secret `%s/%s`: %w",
			namespace,
			secretName,
			err,
		)
	}

	var annotation string
	var draft v1.Secret
	const prelude = `# Please edit the object below. Lines beginning with a '#' will be ignored,
# and an empty file will abort the edit. If an error occurs while saving, this
# file will be reopened with the relevant failures.
#
`
	const limit = 3
	for i := 0; i < limit; i++ {
		// clear the file before writing--this isn't necessary for the first
		// loop iteration, but it's necessary for subsequent attempts.
		if err := tmpFile.Truncate(0); err != nil {
			return fmt.Errorf(
				"editing secret `%s/%s`: seeking file `%s` in advance of "+
					"writing: %w",
				namespace,
				secretName,
				tmpFilePath,
				err,
			)
		}

		// seek to the beginning of the file before writing the file. this is
		// necessary even though we just called `Truncate()` because the latter
		// only changes the file size; not the cursor position.
		if _, err := tmpFile.Seek(0, 0); err != nil {
			return fmt.Errorf(
				"editing secret `%s/%s`: seeking file `%s` in advance of "+
					"reading: %w",
				namespace,
				secretName,
				tmpFilePath,
				err,
			)
		}

		// write the prelude, the annotation (if any), and the serialized secret
		// to the tmp file.
		if _, err := fmt.Fprintf(
			tmpFile,
			"%s# %s\n%s",
			prelude,
			strings.ReplaceAll(annotation, "\n", "\n# "),
			data,
		); err != nil {
			return fmt.Errorf(
				"editing secret: writing secret `%s/%s` to file `%s`: %w",
				namespace,
				secretName,
				tmpFilePath,
				err,
			)
		}
		annotation = "" // clear the annotation

		// sync the write to disk so it's there for the text editor
		if err := tmpFile.Sync(); err != nil {
			return fmt.Errorf(
				"editing secret `%s/%s`: syncing file `%s` following write: %w",
				namespace,
				secretName,
				tmpFilePath,
				err,
			)
		}

		// open the file in the text editor and wait until the editor program
		// exits
		if err := runTextEditor(ctx, editor, tmpFilePath); err != nil {
			return fmt.Errorf(
				"editing secret `%s/%s`: %w",
				namespace,
				secretName,
				err,
			)
		}

		// read the file contents into `data`. we will unmarshal this data, and
		// if we need to retry the editor (e.g., because yaml unmarshaling
		// failed) we will write this data back to the file at the top of the
		// next loop.
		if err = readFile(tmpFile, &data); err != nil {
			return fmt.Errorf(
				"editing secret `%s/%s`: reading changes from file `%s`: %w",
				namespace,
				secretName,
				tmpFilePath,
				err,
			)
		}

		// unmarshal the file changes into the `draft` secret object. if there
		// is an error, then set the error message as the annotation for the
		// next iteration.
		if err := yaml.Unmarshal(data, &draft); err != nil {
			annotation = err.Error()
			continue
		}

		// if there are no changes, exit early
		if secretEqual(original, &draft) {
			slog.Info("no changes detected")
			return nil
		}

		// otherwise update the secret from the draft. if there is an error,
		// set the error message as the annotation for the next iteration.
		if _, err := secrets.Update(
			ctx,
			&draft,
			metav1.UpdateOptions{},
		); err != nil {
			annotation = err.Error()
			continue
		}

		// if we got here, return success
		return nil
	}

	return fmt.Errorf("giving up after %d attempts", limit)
}

func readFile(file *os.File, data *[]byte) error {
	// seek to the beginning of the file before reading the file into
	// memory.
	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf(
			"reading file contents: seeking file in advance of reading: %w",
			err,
		)
	}
	*data = (*data)[0:0] // reset

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Bytes()
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) > 0 && trimmed[0] == '#' {
			continue
		}

		*data = append(append(*data, line...), '\n')
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading file contents: scanner error: %w", err)
	}
	return nil
}

func runTextEditor(ctx context.Context, editor, filePath string) error {
	cmd := exec.CommandContext(ctx, editor, filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: editing file `%s`: %w", editor, filePath, err)
	}
	return nil
}

const tmpDir = "/tmp/kube-edit-secret"
