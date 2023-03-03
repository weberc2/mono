package agent

import (
	"io"
)

type PrefixWriter struct {
	w             io.Writer
	prefix        []byte
	buffer        []byte
	writtenBefore bool
}

func NewPrefixWriter(w io.Writer, prefix string) *PrefixWriter {
	return &PrefixWriter{w: w, prefix: []byte(prefix)}
}

func (pw *PrefixWriter) Write(data []byte) (int, error) {
	pw.buffer = pw.buffer[0:0]
	if !pw.writtenBefore {
		pw.buffer = append(pw.buffer, pw.prefix...)
		pw.writtenBefore = true
	}
	for _, b := range data {
		pw.buffer = append(pw.buffer, b)
		if b == '\n' {
			pw.buffer = append(pw.buffer, pw.prefix...)
		}
	}
	n, err := pw.w.Write(pw.buffer)

	// The io.Copy family of functions will error out if more data is written
	// to a writer than was provided.
	// https://github.com/golang/go/blob/b146d7626f869901f9dd841b9253e89a227c6465/src/io/io.go#L430-L435
	if n > len(data) {
		return len(data), err
	}
	return n, err
}
