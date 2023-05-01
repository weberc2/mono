package fs

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestBuffer_SeekBeyondLenGrows(t *testing.T) {
	b := NewBuffer(nil)
	b.Seek(512, 0)
	if wanted, found := 512, b.Len(); wanted != found {
		t.Fatalf("Len(): wanted `%d`; found `%d`", wanted, found)
	}
}

func TestBuffer_SeekWrite(t *testing.T) {
	b := NewBuffer([]byte{0, 0})
	n, err := b.Seek(1, io.SeekStart)
	if err != nil {
		t.Fatalf("Seek(): unexpected err: %v", err)
	}
	if n != 1 {
		t.Fatalf("Seek(): wanted `1` byte; found `%d`", n)
	}

	n2, err := b.Write([]byte{1})
	if err != nil {
		t.Fatalf("Write(): unexpected err: %v", err)
	}
	if n2 != 1 {
		t.Fatalf("Write(): wanted `1` byte; found `%d`", n2)
	}

	if wanted, found := []byte{0, 1}, b.Bytes(); !bytes.Equal(wanted, found) {
		t.Fatalf("Write(): wanted `%#x`; found `%#x`", wanted, found)
	}
}

func TestBuffer_Overwrite(t *testing.T) {
	// Given a buffer with some existing data
	b := NewBuffer([]byte{0})

	// When some other data is written (with the same size as the existing
	// data)
	wanted := []byte{1}
	b.Write(wanted)

	// Then the data in the buffer should contain the most recently written
	// data
	if found := b.Bytes(); !bytes.Equal(found, wanted) {
		t.Fatalf(
			"Buffer.Write(twos): wanted `%#x`; found `%#x",
			wanted,
			found,
		)
	}
}

func TestBuffer_OverwritePrefix(t *testing.T) {
	// Given a big buffer
	b := NewBuffer([]byte("hello world"))

	// When some initial data is overwritten
	newPrefix := []byte("bye  ")
	b.Write(newPrefix)

	// Then the data in the buffer should contain the overwritten prefix plus
	// the remaining suffix
	wanted := []byte("bye   world")
	if found := b.Bytes(); !bytes.Equal(found, wanted) {
		t.Fatalf(
			"Buffer.Write(twos): wanted `%s`; found `%s`",
			wanted,
			found,
		)
	}
}

func TestBuffer_WriteGrows(t *testing.T) {
	b := NewBuffer(nil)
	wantedData := []byte{1, 1, 1, 1}
	foundWritten, err := b.Write(wantedData)
	if err != nil {
		// should never get here; write should never return an err
		t.Fatalf("unexpected write err: %v", err)
	}

	if wantedWritten := 4; wantedWritten != foundWritten {
		t.Fatalf(
			"Write(): wanted `%d` bytes written; found `%d`",
			wantedWritten,
			foundWritten,
		)
	}

	if foundData := b.Bytes(); !bytes.Equal(wantedData, foundData) {
		t.Fatalf("Bytes(): wanted `%#x`; found `%#x`", wantedData, foundData)
	}
}

func TestBuffer_Read(t *testing.T) {
	wantedData := []byte{1, 2, 3, 4}
	b := NewBuffer(wantedData)
	foundData := [4]byte{}
	foundN, err := b.Read(foundData[:])
	if err != nil {
		t.Fatalf("unexpected read err: %v", err)
	}

	if foundN != len(foundData) {
		t.Fatalf(
			"Read(): wanted `%d` bytes read; found `%d`",
			len(foundData),
			foundN,
		)
	}

	if !bytes.Equal(wantedData, foundData[:foundN]) {
		t.Fatalf("Read(): wanted `%#x`; found `%#x`", wantedData, foundData)
	}
}

func TestBuffer_ReadEOF(t *testing.T) {
	wantedData := []byte{1}
	foundData := [4]byte{}
	b := NewBuffer(wantedData)
	foundN, err := b.Read(foundData[:])
	if err != nil {
		t.Fatalf("Read(): unexpected err: %v", err)
	}

	if wantedN := len(wantedData); foundN != 1 {
		t.Fatalf(
			"Read(): expected `%d` bytes read; found `%d`",
			wantedN,
			foundN,
		)
	}

	if !bytes.Equal(wantedData, foundData[:foundN]) {
		t.Fatalf("Read(): wanted `%#x`; found `%#x`", wantedData, foundData)
	}

	// Reading a second time should return io.EOF
	wantedData = make([]byte, 4)
	foundData = [4]byte{}
	foundN, err = b.Read(foundData[:])
	if !errors.Is(io.EOF, err) {
		t.Fatalf("Read(): expected err `io.EOF`; found `%v`", err)
	}

	if foundN != 0 {
		t.Fatalf(
			"Read() (2nd time): expected `0` bytes read; found `%d`",
			foundN,
		)
	}

	if !bytes.Equal(wantedData, foundData[:]) {
		t.Fatalf(
			"Read() (2nd time): expected `%#x`; found `%#x`",
			wantedData,
			foundData,
		)
	}
}
