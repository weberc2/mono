package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash/adler32"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Architecture string

const (
	ArchitectureAMD64 Architecture = "amd64"
	ArchitectureARM64 Architecture = "arm64"
)

func (a Architecture) Validate() error {
	switch a {
	case ArchitectureAMD64:
		return nil
	case ArchitectureARM64:
		return nil
	default:
		return fmt.Errorf("invalid architecture: %s", a)
	}
}

func (a Architecture) ToAWS() *string {
	switch a {
	case ArchitectureAMD64:
		return aws.String("x86_64")
	case ArchitectureARM64:
		return aws.String("arm64")
	default:
		panic(fmt.Sprintf("invalid architecture: %s", a))
	}
}

type Payload struct {
	// Name is the name to give to the created lambda.
	Name string `json:"name"`

	// Architecture is the target architecture for compilation. Defaults to
	// `arm64`.
	Architecture Architecture `json:"architecture"`

	// Archive contains the base64-encoded zip data containing the source code.
	Archive string `json:"archive"`
}

func (p *Payload) Validate() error {
	if strings.Contains(p.Name, "/") {
		return fmt.Errorf("validating payload: invalid name: `%s`", p.Name)
	}

	if p.Architecture == "" {
		p.Architecture = ArchitectureARM64
	}

	return nil
}

type Builder struct {
	Client *s3.S3
	Bucket string
	Prefix string
}

type Response struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
	Hash   string `json:"hash"`
}

func (builder *Builder) Build(
	ctx context.Context,
	payload *Payload,
) (*Response, error) {
	log.WithField("payload", payload).Infof("building Go project")

	if err := payload.Validate(); err != nil {
		return nil, fmt.Errorf("building lambda: %w", err)
	}
	log.Infof("validated payload")

	// provision temporary working directory
	tmpDir := os.TempDir()
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			log.Errorf(
				"building lambda `%s`: cleaning up temporary directory `%s`:"+
					"%v",
				payload.Name,
				tmpDir,
				err,
			)
		}
	}()
	log.Infof("created temporary directory")

	// extract the source archive to the temporary working directory
	if err := extract(payload.Archive, tmpDir); err != nil {
		return nil, fmt.Errorf("building lambda `%s`: %w", payload.Name, err)
	}
	log.Infof("extracted source archive")

	// run the build inside the temporary working directory
	if err := compileGoProject(tmpDir, payload.Architecture); err != nil {
		return nil, fmt.Errorf("building lambda `%s`: %w", payload.Name, err)
	}
	log.Infof("compiled source code")

	// zip the executable
	zipped, err := mkzip(tmpDir, "main")
	if err != nil {
		return nil, fmt.Errorf("building lambda `%s`: %w", payload.Name, err)
	}
	log.Infof("zipped binary executable")

	// push the zipped executable to s3
	checksum := checksum(zipped)
	key, err := builder.pushToS3(
		ctx,
		payload.Name,
		payload.Architecture,
		zipped,
		checksum,
	)
	if err != nil {
		return nil, fmt.Errorf("building lambda `%s`: %w", payload.Name, err)
	}
	log.Infof("pushed zipped binary to s3://%s/%s", builder.Bucket, key)

	return &Response{Bucket: builder.Bucket, Key: key, Hash: checksum}, nil
}

func compileGoProject(dir string, goarch Architecture) error {
	cmd := exec.Command("go", "build", "-v", "-o", "main")
	cmd.Env = append(
		cmd.Env,
		"GOOS=linux",
		fmt.Sprintf("GOARCH=%s", goarch),
		fmt.Sprintf("GOPATH=%s", os.Getenv("GOPATH")),
		fmt.Sprintf("GOMODCACHE=%s", "/tmp/gomodcache"),
		fmt.Sprintf("GOCACHE=%s", "/tmp/gocache"),
		fmt.Sprintf("HOME=%s", os.Getenv("HOME")),
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		log.WithField("err", err.Error()).Infof("build failed")
		return fmt.Errorf("compiling Go project: %w", err)
	}

	return nil
}

func (builder *Builder) pushToS3(
	ctx context.Context,
	payloadName string,
	architecture Architecture,
	data []byte,
	checksum string,
) (string, error) {
	key := path.Join(
		builder.Prefix,
		payloadName,
		string(architecture),
		fmt.Sprintf("%s.zip", checksum),
	)

	if _, err := builder.Client.PutObjectWithContext(
		ctx,
		&s3.PutObjectInput{
			Bucket: &builder.Bucket,
			Key:    &key,
			Body:   bytes.NewReader(data),
		},
	); err != nil {
		return "", fmt.Errorf(
			"pushing zip to s3://%s/%s: %w",
			builder.Bucket,
			key,
			err,
		)
	}

	return key, nil
}

func checksum(data []byte) string {
	hash := adler32.New()
	_, _ = hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}

func mkzip(exeDir, exeName string) ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.CreateHeader(&zip.FileHeader{
		Name:    exeName,
		NonUTF8: true,
	})
	if err != nil {
		return nil, fmt.Errorf(
			"building lambda zip from executable: creating header for "+
				"executable: %w",
			err,
		)
	}

	path := filepath.Join(exeDir, exeName)
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf(
			"building lambda zip from executable: opening executable file: %w",
			err,
		)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Errorf("closing file `%s`: %v", path, err)
		}
	}()

	if _, err := io.Copy(w, f); err != nil {
		return nil, fmt.Errorf(
			"building lambda zip from executable: copying executable into "+
				"zip file: %w",
			err,
		)
	}

	if err := zw.Flush(); err != nil {
		return nil, fmt.Errorf(
			"building lambda zip from executable: flushing zip writer: %w",
			err,
		)
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf(
			"building lambda zip from executable: closing zip writer: %w",
			err,
		)
	}

	return buf.Bytes(), nil
}

func extract(archive string, dir string) error {
	log.WithField("archive", string(archive)).Infof("extracting archive")
	data, err := base64.StdEncoding.DecodeString(archive)
	if err != nil {
		return fmt.Errorf("base64-decoding archive: %w", err)
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("constructing zip reader: %w", err)
	}

	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(
				filepath.Join(dir, f.Name),
				0755,
			); err != nil {
				return fmt.Errorf(
					"extracting archive: creating directory `%s` inside "+
						"directory `%s`: %w",
					f.Name,
					dir,
					err,
				)
			}
		} else {
			path := filepath.Join(dir, f.Name)
			if err := func() error {
				dst, err := os.Create(path)
				if err != nil {
					return fmt.Errorf("opening dst file `%s`: %w", path, err)
				}
				defer func() {
					if err := dst.Close(); err != nil {
						log.Errorf("closing dst file `%s`: %v", path, err)
					}
				}()

				src, err := f.Open()
				if err != nil {
					return fmt.Errorf(
						"opening src file `%s` from archive: %w",
						f.Name,
						err,
					)
				}
				defer func() {
					if err := src.Close(); err != nil {
						log.Errorf("closing archive file `%s`: %v", path, err)
					}
				}()

				if _, err := io.Copy(dst, src); err != nil {
					return fmt.Errorf("writing to `%s`: %w", path, err)
				}
				return nil
			}(); err != nil {
				return fmt.Errorf(
					"extracting zip file: extracting file `%s`: %w",
					f.Name,
					err,
				)
			}
		}
	}

	return nil
}
