package main

import (
	"log"
	"os"

	"github.com/weberc2/mono/pkg/pgutil/cli"
	"github.com/weberc2/mono/pkg/pgcommentsstore"
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
