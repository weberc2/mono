package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type BuildDestinationS3 struct {
	Region string `json:"region"`
	Bucket Bucket `json:"bucket"`
	Prefix string `json:"prefix"`
}

func (bds3 *BuildDestinationS3) Upload(
	ctx context.Context,
	secrets *Secrets,
	directory string,
	output BuildOutput,
) error {
	creds, found := secrets.S3BucketCredentials[bds3.Bucket]
	if !found {
		return fmt.Errorf(
			"uploading artifact `%s` to `s3://%s/%s/%s`: "+
				"no credentials for bucket: %s",
			output,
			bds3.Bucket,
			bds3.Prefix,
			output,
			bds3.Bucket,
		)
	}

	sess, err := session.NewSession(&aws.Config{
		Region: &bds3.Region,
		Credentials: credentials.NewStaticCredentialsFromCreds(
			credentials.Value{
				AccessKeyID:     creds.AccessKeyID,
				SecretAccessKey: creds.SecretAccessKey,
			},
		),
	})
	if err != nil {
		return fmt.Errorf(
			"uploading artifact to `s3://%s/%s/%s`: "+
				"loading AWS credentials for bucket: %w",
			bds3.Bucket,
			bds3.Prefix,
			output,
			err,
		)
	}

	path := filepath.Join(directory, string(output))
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf(
			"uploading artifact to `s3://%s/%s/%s`: "+
				"opening file `%s`: %w",
			bds3.Bucket,
			bds3.Prefix,
			output,
			path,
			err,
		)
	}

	defer func() {
		if err := file.Close(); err != nil {
			slog.Error(
				"uploading artifact to s3; closing artifact file",
				"err", err.Error(),
				"bucket", bds3.Bucket,
				"prefix", bds3.Prefix,
				"artifact", output,
			)
		}
	}()

	if _, err := s3.New(sess).PutObject(&s3.PutObjectInput{
		Body:   file,
		Bucket: (*string)(&bds3.Bucket),
		Key:    aws.String(fmt.Sprintf("%s/%s", bds3.Prefix, output)),
	}); err != nil {
		return fmt.Errorf(
			"uploading artifact to `s3://%s/%s/%s`: "+
				"putting object: %w",
			bds3.Bucket,
			bds3.Prefix,
			output,
			err,
		)
	}

	return nil
}

type Bucket string
