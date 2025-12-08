package main

import (
	"context"
	"kubestatus/pkg/kubestatus"
	"log"
)

func main() {
	if err := kubestatus.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
