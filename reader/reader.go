// Package reader provides the utilities to consume a RIFF
// data stream and extract details about it.
package reader

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/standoffvenus/goriffa"
	"github.com/standoffvenus/goriffa/internal"
)

// Reader provides a mechanism for reading RIFF
// data chunks.
type Reader struct {
	fileType  internal.FileType
	size      uint32
	bytesRead int64

	r io.Reader
}

var _ goriffa.Reader = new(Reader)

var (
	errCorruptedTooShort        error = fmt.Errorf("%w: %s", internal.ErrCorrupted, internal.ErrBufferUnderflow)
	errCorruptedReadOutOfBounds error = fmt.Errorf("%w: read outside file size - file must be corrupt", internal.ErrCorrupted)
)

// New will create a new RIFF reader that reads
// the provided io.Reader for RIFF data. If the
// reader does not begin with a valid RIFF header,
// goriffa.ErrCorrupted will be returned.
//
// The returned reader is NOT concurrent-safe.
func New(r io.Reader) (*Reader, error) {
	var (
		riffPrefix internal.FourCC
		size       [4]byte
		fileType   internal.FileType
	)

	if _, err := internal.Read(r,
		riffPrefix[:],
		size[:],
		fileType[:],
	); err != nil {
		return nil, wrap(err)
	}
	if !bytes.Equal(riffPrefix[:], goriffa.FourCCRIFF[:]) {
		return nil, fmt.Errorf("%w: data does not begin with RIFF header", internal.ErrCorrupted)
	}

	parsedSize := binary.LittleEndian.Uint32(size[:])
	riffReader := &Reader{
		fileType: fileType,
		size:     parsedSize,
		r:        r,
	}
	if riffReader.size < 4 {
		return nil, fmt.Errorf("%w: impossibly small file size (%d)", internal.ErrCorrupted, riffReader.size)
	}
	riffReader.bytesRead = int64(len(fileType))

	return riffReader, nil
}

// ReadChunk will read the next chunk from the underlying
// reader. On success, the number of bytes read will be
// returned.
//
// The returned number of bytes will be:
//  (goriffa.LengthChunkHeader + len(c.Data)) + 1 (if there is padding)
//
// If at any point during reads an underflow occurs, ErrCorrupted
// will be returned. If any underlying reader error occurs,
// it will be returned.
func (r *Reader) ReadChunk(chunk *internal.Chunk) (int, error) {
	var header [internal.LengthChunkHeader]byte
	headerN, headerErr := r.read(header[:])
	if headerErr != nil {
		return headerN, headerErr
	} else if headerN < len(header) {
		return headerN, errCorruptedTooShort
	}

	chunkSize := binary.LittleEndian.Uint32(header[4:])
	data := internal.Pad(make([]byte, chunkSize))
	dataN, dataErr := r.read(data)

	totalN := headerN + dataN

	chunk.Identifier = internal.FourCC(internal.Must4Byte(header[:4]))
	chunk.Data = data[:chunkSize] // Padded chunks may contain an extra byte

	if dataErr != nil {
		return totalN, dataErr
	} else if dataN < len(data) {
		return totalN, errCorruptedTooShort
	}

	return totalN, nil
}

// ReadToEnd will call ReadChunk until io.EOF is returned,
// building up a slice of chunks. If any other error is
// returned (i.e. not io.EOF), then all the chunks read
// successfully until error will be returned along with
// the error.
func (r *Reader) ReadToEnd() ([]internal.Chunk, error) {
	// Start with 8 chunks allocated just to avoid too
	// many reallocations.
	chunks := make([]internal.Chunk, 0, 8)
	for {
		var ch internal.Chunk
		if _, err := r.ReadChunk(&ch); err != nil {
			if errors.Is(err, io.EOF) {
				return chunks, nil
			}

			return chunks, err
		}

		chunks = append(chunks, ch)
	}
}

// FileType returns the parsed file type for the RIFF
// data.
func (r *Reader) FileType() internal.FileType {
	return r.fileType
}

// Size returns the content length as reported by
// the RIFF data read.
func (r *Reader) Size() uint32 {
	return r.size
}

func (r *Reader) read(b []byte) (int, error) {
	n, err := r.r.Read(b)
	r.bytesRead += int64(n)
	if err != nil {
		return n, err
	}

	if r.bytesRead > internal.PaddedLength(int64(r.size)) {
		return n, errCorruptedReadOutOfBounds
	}

	return n, nil
}

func wrap(err error) error {
	if errors.Is(err, internal.ErrBufferUnderflow) || errors.Is(err, io.EOF) {
		return errCorruptedTooShort
	}

	return err
}
