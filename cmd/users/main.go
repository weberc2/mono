package main

import (
	"log"
	"os"

	"github.com/weberc2/mono/pkg/pguserstore"
	"github.com/weberc2/mono/pkg/pgutil/cli"
)

func main() {
	app, err := cli.New(&pguserstore.Table)
	if err != nil {
		log.Fatal(err)
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
