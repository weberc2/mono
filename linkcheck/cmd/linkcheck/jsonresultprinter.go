package main

import (
	"encoding/json"
	"fmt"
)

type JSONResultPrinter struct {
	indent bool
}

func (printer *JSONResultPrinter) SetIndent(indent bool) *JSONResultPrinter {
	printer.indent = indent
	return printer
}

func (printer *JSONResultPrinter) Visit(r *Result) { printer.Print(r) }

func (printer *JSONResultPrinter) Print(r *Result) {
	var (
		data []byte
		err  error
	)
	if printer.indent {
		data, err = json.MarshalIndent(r, "", "  ")
	} else {
		data, err = json.Marshal(r)
	}

	if err != nil {
		panic("cannot marshal Result type")
	}

	fmt.Printf("%s\n", data)
}
