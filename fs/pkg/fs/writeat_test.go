package fs

import (
	"bytes"
	"testing"
)

func TestWriteAt(t *testing.T) {
	buf := NewBuffer([]byte{0, 0, 0, 0})
	if err := WriteAt(buf, 2, []byte{1, 1}); err != nil {
		t.Fatalf("WriteAt(): unexpected err: %v", err)
	}

	wanted := []byte{0, 0, 1, 1}
	if !bytes.Equal(wanted, buf.Bytes()) {
		t.Fatalf("WriteAt(): wanted `%#x`; found `%#x`", wanted, buf.Bytes())
	}
}
