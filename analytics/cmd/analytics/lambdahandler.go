package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

func HandleLambda(
	ctx context.Context,
	req events.APIGatewayProxyRequest,
) (rsp events.APIGatewayProxyResponse, err error) {
	slog.Info("received event", "event", &req)
	now := time.Now().UTC()

	var record Record
	record.Path = req.QueryStringParameters["path"]
	record.Host = req.QueryStringParameters["host"]
	record.Proto = req.QueryStringParameters["proto"]
	record.Referer = req.QueryStringParameters["referer"]
	record.UserAgent = req.RequestContext.Identity.UserAgent
	record.SourceIP = req.RequestContext.Identity.SourceIP
	record.Time = now

	if record.Location, err = client.Locate(ctx, record.SourceIP); err != nil {
		slog.Error("locating ip", "err", err.Error(), "ip", record.SourceIP)
	}

	y, M, d := now.Date()
	h, m, s := now.Clock()
	us := now.Nanosecond() / 1000
	key := fmt.Sprintf("%d/%02d/%02d/%02d/%02d/%02d.%06d", y, M, d, h, m, s, us)

	data, err := json.Marshal(&record)
	if err != nil {
		slog.Error(
			"marshaling record to json",
			"err", err.Error(),
			"record", &record,
		)
	}

	slog.Info("Inserting data into s3", "bucket", bucket, "key", key, "data", string(data))

	// TODO write to s3 & return output
	return {
        "isBase64Encoded": False,
        "statusCode": 200,
        "headers": {},
        "multiValueHeaders": {},
        "body": ""
    }
}

type Record struct {
	Path      string    `json:"path"`
	Host      string    `json:"host"`
	Proto     string    `json:"proto"`
	Referer   string    `json:"referer"`
	UserAgent string    `json:"user_agent"`
	SourceIP  string    `json:"source_ip"`
	Time      time.Time `json:"time"`
	Location
}

var (
	bucket = os.Getenv("BUCKET")
	client = Client{HTTP: &http.Client{Timeout: 15 * time.Second}}
)

func init() {
	secret := os.Getenv("SECRET")
	sess, err := session.NewSession(aws.NewConfig())
	if err != nil {
		log.Fatal(err)
	}
	svc := secretsmanager.New(sess)
	rsp, err := svc.GetSecretValueWithContext(
		aws.BackgroundContext(),
		&secretsmanager.GetSecretValueInput{SecretId: &secret},
	)
	if err != nil {
		log.Fatal(err)
	}
	client.APIKey = *rsp.SecretString
}
