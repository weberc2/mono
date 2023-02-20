package main

import (
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	bucketPrefix, exists := os.LookupEnv("BUCKET_PREFIX")
	if !exists {
		bucketPrefix = "gobuilder"
	}

	builder := &Builder{
		Client: s3.New(session.Must(session.NewSession())),
		Bucket: os.Getenv("BUCKET"),
		Prefix: bucketPrefix,
	}

	if builder.Bucket == "" {
		logrus.Fatalf("missing required env var: `BUCKET`")
	}

	logrus.WithField("bucket", builder.Bucket).
		WithField("prefix", builder.Prefix).
		Infof("starting lambda function")
	lambda.Start(builder.Build)
}
