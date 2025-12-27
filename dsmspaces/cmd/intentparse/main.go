package main

import (
	"bufio"
	"bytes"
	"context"
	"dsmspaces/pkg/dsmspaces"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	apiKey := os.Getenv("OPENAIAPIKEY")
	if apiKey == "" {
		return fmt.Errorf("missing required environment variable: OPENAIAPIKEY")
	}
	var (
		parser  = dsmspaces.NewIntentsParser(apiKey)
		scanner = bufio.NewScanner(os.Stdin)
		buf     bytes.Buffer
	)
	for {
		fmt.Printf(" > ")
		if !scanner.Scan() {
			return scanner.Err()
		}

		intents, err := parser.ParseIntentsJSON(ctx, scanner.Text())
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			// we will continue to print the intents json for debugging purposes
		}

		buf.Reset()
		if err = json.Indent(&buf, intents, "", "  "); err != nil {
			fmt.Fprintf(os.Stderr, "formatting json: %v\n", err)
		} else {
			intents = buf.Bytes()
		}

		fmt.Printf("%s\n", intents)
	}
}
