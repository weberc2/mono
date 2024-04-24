package main

import (
	"log"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	svc, err := LoadService()
	if err != nil {
		log.Fatal(err)
	}
	lambda.Start(svc.Handle)
}
