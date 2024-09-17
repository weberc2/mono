package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	start := time.Now()
	slog.Debug("starting", "time", start)
	defer func() { slog.Debug("completed", "elapsed", time.Since(start)) }()

	repo := "git@github.com:weberc2/mono.git"
	key, err := ssh.NewPublicKeysFromFile(
		"git",
		"/Users/weberc2/.ssh/id_ecdsa",
		"",
	)
	if err != nil {
		log.Fatalf("loading private key: %v", err)
	}

	var bucket Bucket = "weberc2"

	service := Service{
		Secrets: &Secrets{
			RepositoryAuthMethods: map[string]transport.AuthMethod{repo: key},
			S3BucketCredentials: map[Bucket]*AWSCredentials{
				bucket: {
					AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
					SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
				},
			},
		},
		BuildParentDirectory: "/tmp",
	}

	if err := service.Build(
		context.Background(),
		&Build{
			Context: BuildContext{
				Type: BuildContextTypeGit,
				Backend: &BuildContextGit{
					Repository: repo,
					Branch:     "master",
					Directory:  "gobuilder",
				},
			},
			Package: "./cmd/gobuilder",
			Output:  "gobuilder",
			Destination: BuildDestination{
				Type: BuildDestinationTypeS3,
				Backend: &BuildDestinationS3{
					Region: "us-east-1",
					Bucket: bucket,
					Prefix: "delete-me",
				},
			},
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Stdin:  os.Stdin,
		},
	); err != nil {
		log.Fatal(err)
	}
}
