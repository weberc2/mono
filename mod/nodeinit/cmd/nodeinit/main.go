package main

import (
	"context"
	"log"
	"os"

	"github.com/weberc2/mono/mod/nodeinit/pkg/agent"
)

func main() {
	agent := agent.New()

	if serverAddr := os.Getenv("NODEINIT_SERVER_ADDR"); serverAddr != "" {
		agent.Client.SetServerAddr(serverAddr)
	}

	if err := agent.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
