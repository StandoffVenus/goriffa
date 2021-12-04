package test

import (
	"bytes"
	"io"

	"github.com/standoffvenus/goriffa/internal"
)

// Buffer extends bytes.Buffer to implement
// io.WriterAt.
type Buffer struct {
	bytes.Buffer
}

var _ io.WriterAt = new(Buffer)

func (buf *Buffer) WriteAt(b []byte, offset int64) (int, error) {
	currentBytes := buf.Bytes()
	if offset > int64(len(currentBytes)) {
		internal.Panic("offset out of bounds")
	}

	n := copy(currentBytes[offset:], b)
	if n < len(b) {
		currentBytes = append(currentBytes, b[n:]...)
	}

	buf.Buffer = *bytes.NewBuffer(currentBytes)

	return len(b), nil
}
