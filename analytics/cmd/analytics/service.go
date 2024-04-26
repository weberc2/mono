package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
	"unsafe"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

// Service handles requests. For each request, it stores analytics information
// in s3, including location information for the source IP address.
type Service struct {
	// S3 is the s3 client.
	S3 *s3.S3

	// Bucket is the s3 bucket where analytics information is written.
	Bucket string

	// Client is the HTTP client used to look up location information from
	// locator services.
	Client http.Client

	// Locator abstracts over different locator services.
	Locator MultiLocator
}

// LoadService loads a `Service` from the environment.
func LoadService() (svc Service, err error) {
	sess, err := session.NewSession(aws.NewConfig())
	if err != nil {
		err = fmt.Errorf("loading service from environment: %w", err)
		return
	}

	secret := os.Getenv("SECRET")
	var rsp *secretsmanager.GetSecretValueOutput
	if rsp, err = secretsmanager.New(sess).GetSecretValueWithContext(
		aws.BackgroundContext(),
		&secretsmanager.GetSecretValueInput{SecretId: &secret},
	); err != nil {
		err = fmt.Errorf("loading service from environment: %w", err)
		return
	}

	svc = Service{
		S3:     s3.New(sess),
		Bucket: os.Getenv("BUCKET"),
		Client: http.Client{Timeout: 15 * time.Second},
	}

	if err = json.Unmarshal(
		*(*[]byte)(unsafe.Pointer(rsp.SecretString)),
		&svc.Locator.Locators,
	); err != nil {
		err = fmt.Errorf(
			"loading service from environment: unmarshaling locators from "+
				"secret string: %w",
			err,
		)
		return
	}

	if len(svc.Locator.Locators) < 1 {
		err = fmt.Errorf(
			"loading service from environment: no locators found in secret",
		)
		return
	}

	return
}

// Handle handles the API Gateway request. It stores request data and attempts
// to look up the location of the source IP address.
func (svc *Service) Handle(
	ctx context.Context,
	req events.APIGatewayV2HTTPRequest,
) (rsp events.APIGatewayV2HTTPResponse, err error) {
	// always return OK--users don't need to see failed requests
	rsp.StatusCode = http.StatusOK

	const keyFmt = "%d/%02d/%02d/%02d/%02d/%02d.%06d"
	var (
		now     = time.Now().UTC()
		y, M, d = now.Date()
		h, m, s = now.Clock()
		us      = now.Nanosecond() / 1000
		key     = fmt.Sprintf(keyFmt, y, M, d, h, m, s, us)
		data    []byte
		r       struct {
			Path      string    `json:"path"`
			Host      string    `json:"host"`
			Proto     string    `json:"proto"`
			Referer   string    `json:"referer"`
			UserAgent string    `json:"user_agent"`
			SourceIP  string    `json:"source_ip"`
			Time      time.Time `json:"time"`
			Location
		}
	)
	slog.Info("received event", "event", &req)

	r.Path = req.QueryStringParameters["path"]
	r.Host = req.QueryStringParameters["host"]
	r.Proto = req.QueryStringParameters["proto"]
	r.Referer = req.QueryStringParameters["referer"]
	r.UserAgent = req.RequestContext.HTTP.UserAgent
	r.SourceIP = req.RequestContext.HTTP.SourceIP
	r.Time = now

	// try to write the record to s3 even if there is a problem fetching
	// location information.
	defer func() {
		if data, err = json.Marshal(&r); err != nil {
			slog.Error(
				"marshaling record to json",
				"err", err.Error(),
				"record", &r,
			)
			return
		}

		slog.Info(
			"inserting data into s3",
			"bucket", svc.Bucket,
			"key", key,
			"data", json.RawMessage(data),
		)
		if _, err = svc.S3.PutObjectWithContext(
			ctx,
			&s3.PutObjectInput{
				Body:          bytes.NewReader(data),
				Bucket:        &svc.Bucket,
				ContentLength: aws.Int64(int64(len(data))),
				ContentType:   aws.String("application/json"),
				Key:           &key,
			},
		); err != nil {
			slog.Error("inserting data into s3", "err", err.Error())
		}
	}()

	if r.Location, err = svc.Locator.Locate(
		ctx,
		&svc.Client,
		r.SourceIP,
	); err != nil {
		slog.Error("locating ip", "err", err.Error(), "ip", r.SourceIP)
	}

	return
}
