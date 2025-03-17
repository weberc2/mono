package main

import "github.com/go-git/go-git/v5/plumbing/transport"

type Secrets struct {
	RepositoryAuthMethods map[string]transport.AuthMethod `json:"repositoryAuthMethods"`
	S3BucketCredentials   map[Bucket]*AWSCredentials      `json:"s3BucketCredentials"`
}

type AWSCredentials struct {
	AccessKeyID     string `json:"accessKeyID"`
	SecretAccessKey string `json:"secretAccessKey"`
}
