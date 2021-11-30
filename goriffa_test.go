package goriffa_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/standoffvenus/goriffa"
	"github.com/standoffvenus/goriffa/internal/test"
	"github.com/standoffvenus/goriffa/reader"
	"github.com/standoffvenus/goriffa/writer"
	"github.com/stretchr/testify/assert"
)

// This file holds the overall "integration" tests for
// Goriffa - that is, verifying real world examples.

func TestWAVE(t *testing.T) {
	r, d := test.WAV()
	testFile(t, r, d)
}

func TestWEBP(t *testing.T) {
	r, d := test.WEBP()
	testFile(t, r, d)
}

func TestWriteReadFile(t *testing.T) {
	tmpFile, tmpFileErr := os.CreateTemp(os.TempDir(), "")
	assert.NoError(t, tmpFileErr)
	defer func() { _ = tmpFile.Close() }()

	w, wErr := writer.New(tmpFile, test.FileType)
	defer func() { _ = w.Close() }()
	assert.NoError(t, wErr)

	chunk := goriffa.Chunk{
		Identifier: goriffa.FourCCData,
		Data:       []byte{3, 2, 1},
	}
	n, writeErr := w.WriteChunk(chunk)
	assert.Equal(t, int64(n), chunk.ByteLength())
	assert.NoError(t, writeErr)
	assert.NoError(t, w.Close()) // Finalize data.

	// Rewind to beginning of file
	_, seekErr := tmpFile.Seek(0, io.SeekStart)
	assert.NoError(t, seekErr)

	r, rErr := reader.New(tmpFile)
	assert.NoError(t, rErr)

	chunks, readErr := r.ReadToEnd()
	assert.NoError(t, readErr)
	assert.Len(t, chunks, 1)

	assert.Equal(t, chunk, chunks[0])
	assert.Equal(t, test.FileType, r.FileType())
	assert.Equal(t,
		uint32(len(test.FileType)+int(chunk.ByteLength())),
		r.Size())
}

func testFile(t *testing.T, r io.Reader, details test.Details) {
	var originalBuffer bytes.Buffer
	teedReader := io.TeeReader(r, &originalBuffer)

	reader, readNewErr := reader.New(teedReader)
	assert.NoError(t, readNewErr)
	assert.Equal(t, details.FileType(), reader.FileType())
	assert.Equal(t, details.Size(), reader.Size())

	var copyBuffer test.Buffer
	writer, writeNewErr := writer.New(&copyBuffer, details.FileType())
	assert.NoError(t, writeNewErr)
	defer func() { _ = writer.Close() }() // In case we panic out, still want to close out the resources

	chunks, err := reader.ReadToEnd()
	assert.NoError(t, err)

	var total uint32 = 0
	for _, c := range chunks {
		written, writeErr := writer.WriteChunk(c)
		assert.NoError(t, writeErr)

		total += uint32(written)
	}
	assert.NoError(t, writer.Close()) // Ensure the data is finalized

LOOP:
	for idx, b := range originalBuffer.Bytes() {
		if b != copyBuffer.Bytes()[idx] {
			fmt.Println(idx, b, copyBuffer.Bytes()[idx])
			break LOOP
		}
	}

	assert.Equal(t, details.Size()-4, uint32(total))
	assert.True(t, bytes.Equal(originalBuffer.Bytes(), copyBuffer.Bytes()))
}
