package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	var lvl slog.Level
	_ = lvl.UnmarshalText([]byte(os.Getenv("LOG_LEVEL")))
	slog.SetDefault(slog.New(slog.NewJSONHandler(
		os.Stderr,
		&slog.HandlerOptions{Level: lvl},
	)))

	svc, err := LoadService()
	if err != nil {
		log.Fatal(err)
	}
	lambda.Start(svc.Handle)
}
