package main

import (
	"log"
	"os"

	"github.com/weberc2/mono/comments/pkg/pgcommentsstore"
	"github.com/weberc2/mono/comments/pkg/pgutil/cli"
)

func main() {
	app, err := cli.New(&pgcommentsstore.Table)
	if err != nil {
		log.Fatal(err)
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
