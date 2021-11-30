// Package writer provides a mechanism for writing
// RIFF data to a stream (io.Writer).
package writer

import (
	"fmt"
	"io"
	"math"

	"github.com/standoffvenus/goriffa"
	"github.com/standoffvenus/goriffa/internal"
)

// WriterWithWriterAt combines the io.Writer and
// io.WriterAt interfaces.
type WriterWithWriterAt interface {
	io.Writer
	io.WriterAt
}

// Writer provides a mechanism for writing RIFF chunks
// to an io.Writer.
type Writer struct {
	w WriterWithWriterAt

	fileType internal.FileType
	fileSize int64
	closed   bool
}

var _ goriffa.Writer = new(Writer)

// New creates a new RIFF writer, initially writing
//  "RIFF", [4 empty bytes for file size], fileType
// to the provided writer, where fileType is the provided
// file type name. The provided writer is expected to
// be at the start of its internal stream. When you
// are done using the writer, call Close() to ensure
// the RIFF data is finalized. Failing to do so will
// lead to corrupted data since writing the RIFF data
// length is deferred until close.
//
// The returned writer is NOT concurrent-safe.
func New(w WriterWithWriterAt, fileType internal.FileType) (*Writer, error) {
	writer := &Writer{
		w:        w,
		fileType: fileType,
	}
	if err := writer.init(); err != nil {
		return nil, err
	}

	return writer, nil
}

// WriteChunk will write the given chunk to the stream
// in the RIFF format:
//   identifier, size, data...
// where "identifier" will be the FOURCC identifier,
// size will be a uint32 holding the length of the
// chunk data, and data will be the byte payload.
// The number of bytes written to the stream and
// an error (if any occurred) will be returned.
//
// The number of bytes written on success will be:
//  (goriffa.LengthChunkHeader + len(c.Data)) + 1 (if padding byte added)
//
// If the writer has written too many bytes (more than
// 4 GB), goriffa.ErrCorrupted will be returned.
// If the writer is closed, goriffa.ErrClosed will be
// returned.
func (w *Writer) WriteChunk(c internal.Chunk) (int, error) {
	if !w.closed {
		// This is an overflow check
		newSize := w.fileSize + c.ByteLength()
		if newSize > math.MaxUint32 || newSize < w.fileSize {
			return 0, fmt.Errorf("%w: wrote too many bytes - size overflow", internal.ErrCorrupted)
		}

		b := internal.Pad(c.Data)
		n, err := internal.Write(
			w.w,
			c.Identifier[:],
			internal.LittleEndianUInt32Bytes(uint32(len(c.Data))),
			b)
		if err != nil {
			return int(n), err
		}

		w.fileSize += n

		return int(n), nil
	}

	return 0, internal.ErrClosed
}

// Close will close the writer, writing the content
// length at the file size offset.
// All seeks and writes after close will fail.
// If the write fails, the write error will be
// returned.
//
// If the writer is already closed, goriffa.ErrClosed
// will be returned.
func (w *Writer) Close() error {
	if !w.closed {
		w.closed = true

		fileSizeBytes := internal.LittleEndianUInt32Bytes(uint32(w.fileSize))
		if _, err := internal.WriteAt(w.w, fileSizeBytes[:], int64(len(goriffa.FourCCRIFF))); err != nil {
			return err
		}

		return nil
	}

	return internal.ErrClosed
}

func (w *Writer) init() error {
	if _, err := internal.Write(w.w,
		goriffa.FourCCRIFF[:],
		internal.EmptyBytes[:],
		w.fileType[:],
	); err != nil {
		return err
	}
	w.fileSize = 4 // Because we wrote the file type

	return nil
}
