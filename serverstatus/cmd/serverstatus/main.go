package main

import (
	"context"
	"log"
	"serverstatus/pkg/serverstatus"
)

func main() {
	if err := serverstatus.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
