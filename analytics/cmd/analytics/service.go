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

type Service struct {
	S3     *s3.S3
	Client MultiClient
	Bucket string
}

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

	svc = Service{S3: s3.New(sess), Bucket: os.Getenv("BUCKET")}

	var locators []struct {
		Type   LocationSource `json:"type"`
		User   string         `json:"user"`
		APIKey string         `json:"apiKey"`
	}
	if err = json.Unmarshal(
		*(*[]byte)(unsafe.Pointer(rsp.SecretString)),
		&locators,
	); err != nil {
		err = fmt.Errorf(
			"loading service from environment: unmarshaling clients from "+
				"secret string: %w",
			err,
		)
		return
	}

	if len(locators) < 1 {
		err = fmt.Errorf(
			"loading service from environment: no api keys found in secret",
		)
		return
	}

	httpClient := http.Client{Timeout: 15 * time.Second}
	svc.Client.Clients = make([]NamedClient, len(locators))
	for i := range locators {
		locator, ok := locatorsBySource[locators[i].Type]
		if !ok {
			err = fmt.Errorf("invalid locator type: %s", locators[i].Type)
			return
		}
		svc.Client.Clients[i] = NamedClient{
			Name: fmt.Sprintf("%s|%s", locators[i].User, locators[i].Type),
			Client: Client{
				HTTP:    &httpClient,
				APIKey:  locators[i].APIKey,
				Locator: locator,
			},
		}
	}
	return
}

func (svc *Service) Handle(
	ctx context.Context,
	req events.APIGatewayProxyRequest,
) (rsp events.APIGatewayProxyResponse, err error) {
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
	r.UserAgent = req.RequestContext.Identity.UserAgent
	r.SourceIP = req.RequestContext.Identity.SourceIP
	r.Time = now

	if r.Location, err = svc.Client.Locate(ctx, r.SourceIP); err != nil {
		slog.Error("locating ip", "err", err.Error(), "ip", r.SourceIP)
		rsp.StatusCode = http.StatusInternalServerError
		return
	}

	if data, err = json.Marshal(&r); err != nil {
		slog.Error(
			"marshaling record to json",
			"err", err.Error(),
			"record", &r,
		)
		rsp.StatusCode = http.StatusInternalServerError
		return
	}

	slog.Info(
		"inserting data into s3",
		"bucket", svc.Bucket,
		"key", key,
		"data", string(data),
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
		rsp.StatusCode = http.StatusInternalServerError
		return
	}

	rsp.StatusCode = http.StatusOK
	return
}
