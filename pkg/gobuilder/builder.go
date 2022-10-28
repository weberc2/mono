package golanglambdabuilder

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/hex"
	"fmt"
	"hash/adler32"
	"io"
	"io/ioutil"
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

	// Archive contains the tar.gz data containing the source code.
	Archive []byte `json:"archive"`
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

func (builder *Builder) Build(ctx context.Context, payload *Payload) error {
	if err := payload.Validate(); err != nil {
		return fmt.Errorf("building lambda: %w", err)
	}
	log.Infof("validated payload")

	// provision temporary working directory
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return fmt.Errorf(
			"building lambda `%s`: creating temporary directory: %w",
			payload.Name,
			err,
		)
	}
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
		return fmt.Errorf("building lambda `%s`: %w", payload.Name, err)
	}
	log.Infof("extracted source archive")

	// run the build inside the temporary working directory
	if err := compileGoProject(tmpDir, payload.Architecture); err != nil {
		return fmt.Errorf("building lambda `%s`: %w", payload.Name, err)
	}
	log.Infof("compiled source code")

	// zip the executable
	zipped, err := mkzip(tmpDir, "main")
	if err != nil {
		return fmt.Errorf("building lambda `%s`: %w", payload.Name, err)
	}
	log.Infof("zipped binary executable")

	// push the zipped executable to s3
	key, err := builder.pushToS3(ctx, payload.Name, zipped)
	if err != nil {
		return fmt.Errorf("building lambda `%s`: %w", payload.Name, err)
	}
	log.Infof("pushed zipped binary to s3://%s/%s", builder.Bucket, key)

	return nil
}

func compileGoProject(dir string, goarch Architecture) error {
	var buildLogs bytes.Buffer
	cmd := exec.Command("go", "build", "-o", "main")
	cmd.Env = append(
		cmd.Env,
		"GOOS=linux",
		fmt.Sprintf("GOARCH=%s", goarch),
		fmt.Sprintf("GOPATH=%s", os.Getenv("GOPATH")),
		fmt.Sprintf("HOME=%s", os.Getenv("HOME")),
	)
	cmd.Stdout = &buildLogs
	cmd.Stderr = &buildLogs
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		log.WithField("buildLogs", buildLogs.String()).Infof("build failed")
		return fmt.Errorf("compiling Go project: %w", err)
	}

	return nil
}

func (builder *Builder) pushToS3(
	ctx context.Context,
	payloadName string,
	data []byte,
) (string, error) {
	key := path.Join(
		builder.Prefix,
		payloadName,
		fmt.Sprintf("%s.zip", checksum(data)),
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

func extract(archive []byte, dir string) error {
	gzr, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return fmt.Errorf("constructing gzip reader: %w", err)
	}
	r := tar.NewReader(gzr)
	for {
		hdr, err := r.Next()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("fetching next entry in tar file: %w", err)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(
				filepath.Join(dir, hdr.Name),
				0755,
			); err != nil {
				return fmt.Errorf(
					"extracting tarball: creating directory `%s` inside "+
						"directory `%s`: %w",
					hdr.Name,
					dir,
					err,
				)
			}
		case tar.TypeReg:
			path := filepath.Join(dir, hdr.Name)
			if err := func() error {
				f, err := os.Create(path)
				if err != nil {
					return fmt.Errorf("opening file `%s`: %w", path, err)
				}
				defer func() {
					if err := f.Close(); err != nil {
						log.Errorf("closing file `%s`: %v", path, err)
					}
				}()

				if _, err := io.Copy(f, r); err != nil {
					return fmt.Errorf(
						"writing to `%s`: %w",
						path,
						err,
					)
				}
				return nil
			}(); err != nil {
				return fmt.Errorf(
					"extracting tarball: extracting file `%s`: %w",
					hdr.Name,
					err,
				)
			}
		default:
			return fmt.Errorf(
				"extracting tarball: unexpected typeflag `%b` for entry `%s`",
				hdr.Typeflag,
				hdr.Name,
			)
		}
	}
}
