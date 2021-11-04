package main

import (
	"io"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3ObjectStore struct {
	Client *s3.S3
}

func (os *S3ObjectStore) PutObject(bucket, key string, data io.ReadSeeker) error {
	_, err := os.Client.PutObject(&s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   data,
	})
	return err
}

func (os *S3ObjectStore) GetObject(bucket, key string) (io.ReadCloser, error) {
	rsp, err := os.Client.GetObject(&s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		if err, ok := err.(awserr.Error); ok {
			if err.Code() == s3.ErrCodeNoSuchKey {
				return nil, &ObjectNotFoundErr{bucket, key}
			}
		}
		return nil, err
	}
	return rsp.Body, nil
}

func (os *S3ObjectStore) ListObjects(bucket, prefix string) ([]string, error) {
	var keys []string
	err := os.Client.ListObjectsPages(
		&s3.ListObjectsInput{
			Bucket: &bucket,
			Prefix: &prefix,
		},
		func(rsp *s3.ListObjectsOutput, lastPage bool) bool {
			for _, object := range rsp.Contents {
				keys = append(keys, *object.Key)
			}
			return true
		},
	)
	return keys, err
}
