package main

import "fmt"

type PrettyResultPrinter struct {
	lastOkay bool
}

func (printer *PrettyResultPrinter) Visit(r *Result) {
	printer.Print(r)
}

func (printer *PrettyResultPrinter) Print(r *Result) {
	if !printer.lastOkay {
		fmt.Print("\n")
	}
	if r.StatusCode != 200 {
		fmt.Printf(
			"\n⛔️ %s <a href=\"%s\">%s</a>: %d",
			r.BaseURL,
			r.TargetURL,
			r.TargetText,
			r.StatusCode,
		)
		printer.lastOkay = false
		return
	}

	if r.URLParseError != nil {
		fmt.Printf(
			"🙅‍♂️ %s <a href=\"%s\">%s</a>: %v",
			r.BaseURL,
			r.TargetURL,
			r.TargetText,
			r.URLParseError,
		)
		printer.lastOkay = false
		return
	}

	if r.NetworkError != nil {
		fmt.Printf(
			"\n⛔️ %s <a href=\"%s\">%s</a>: %v",
			r.BaseURL,
			r.TargetURL,
			r.TargetText,
			r.NetworkError,
		)
	}

	fmt.Print(".")
	printer.lastOkay = true
}
