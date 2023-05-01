package fs

import (
	"bytes"
	"errors"
	"io"
	"math/rand"
	"testing"
)

func TestReadAt(t *testing.T) {
	var data [1024]byte
	rand.Seed(1)
	rand.Read(data[:])

	var buf [512]byte
	offset := Byte(len(buf))
	if err := ReadAt(
		NewBuffer(data[:]),
		offset,
		buf[:],
	); err != nil {
		t.Fatalf("ReadAt(): unexpected err: %v", err)
	}

	if !bytes.Equal(data[offset:], buf[:]) {
		t.Fatalf("ReadAt(): wanted `%#x`; found `%#x`", data[offset:], buf[:])
	}
}

func TestReadAt_ReturnsEOFWhenOffsetAtEnd(t *testing.T) {
	var data [1024]byte
	var buf [1024]byte
	if err := ReadAt(
		NewBuffer(data[:]),
		Byte(len(data)),
		buf[:],
	); errors.Is(io.EOF, err) {
		t.Fatalf("ReadAt(): expected `io.EOF`; found `%v`", err)
	}
}
