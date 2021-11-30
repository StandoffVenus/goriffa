package writer_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/standoffvenus/goriffa"
	"github.com/standoffvenus/goriffa/internal"
	"github.com/standoffvenus/goriffa/internal/test"
	"github.com/standoffvenus/goriffa/writer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNew(t *testing.T) {
	mockWriter := new(MockWriter)
	expectationsNew(mockWriter)
	expectClose(mockWriter)

	w, err := writer.New(mockWriter, test.FileType)
	assert.NoError(t, err)
	assert.NoError(t, w.Close())

	mockWriter.AssertExpectations(t)
}

func TestNewWriteError(t *testing.T) {
	expectedErr := errors.New("error")

	mockWriter := new(MockWriter)
	mockWriter.
		On("Write", mock.Anything).
		Return(int(0), expectedErr).
		Once()

	_, err := writer.New(mockWriter, test.FileType)
	assert.ErrorIs(t, err, expectedErr)

	mockWriter.AssertExpectations(t)
}

func TestClose(t *testing.T) {
	uint32Bytes := internal.LittleEndianUInt32Bytes(4)

	mockWriter := new(MockWriter)
	expectationsNew(mockWriter)
	mockWriter.
		On("WriteAt", uint32Bytes[:], int64(len(goriffa.FourCCRIFF))).
		Return(len(uint32Bytes), test.NilError).
		Once()

	w, err := writer.New(mockWriter, test.FileType)
	assert.NoError(t, err)

	// Ensure that first close is successful and future
	// operations return ErrClosed
	assert.NoError(t, w.Close())

	_, writeErr := w.WriteChunk(goriffa.Chunk{})
	assert.ErrorIs(t, writeErr, goriffa.ErrClosed)
	assert.ErrorIs(t, w.Close(), goriffa.ErrClosed)

	mockWriter.AssertExpectations(t)
}

func TestCloseWriteAtError(t *testing.T) {
	uint32Bytes := internal.LittleEndianUInt32Bytes(4)

	mockWriter := new(MockWriter)
	expectationsNew(mockWriter)
	mockWriter.
		On("WriteAt", uint32Bytes[:], int64(len(goriffa.FourCCRIFF))).
		Return(int(0), errors.New("error")).
		Once()

	w, err := writer.New(mockWriter, test.FileType)
	assert.NoError(t, err)

	assert.Error(t, w.Close())

	_, writeErr := w.WriteChunk(goriffa.Chunk{})
	assert.ErrorIs(t, writeErr, goriffa.ErrClosed)
	assert.ErrorIs(t, w.Close(), goriffa.ErrClosed)

	mockWriter.AssertExpectations(t)
}

// TestWrite is essentially a test of WriteChunk that
// goes through Write.
func TestWrite(t *testing.T) {
	chunk := goriffa.Chunk{
		Identifier: goriffa.FourCCData,
		Data:       []byte("expected data"),
	}
	paddedData := internal.Pad(chunk.Data)

	expectedStream := goriffa.FourCCRIFF[:]
	expectedStream = append(expectedStream, []byte{byte(2*(len(paddedData)+internal.LengthChunkHeader) + 4), 0, 0, 0}...)
	expectedStream = append(expectedStream, test.FileType[:]...)
	expectedStream = append(expectedStream, chunk.Identifier[:]...)
	expectedStream = append(expectedStream, []byte{byte(len(chunk.Data)), 0, 0, 0}...)
	expectedStream = append(expectedStream, paddedData...)
	expectedStream = append(expectedStream, chunk.Identifier[:]...)
	expectedStream = append(expectedStream, []byte{byte(len(chunk.Data)), 0, 0, 0}...)
	expectedStream = append(expectedStream, paddedData...)

	var buffer test.Buffer
	writer, err := writer.New(&buffer, test.FileType)
	assert.NoError(t, err)

	n, err := writer.WriteChunk(chunk)
	assert.NoError(t, err)
	assert.Equal(t, len(paddedData)+internal.LengthChunkHeader, n)

	n, err = writer.WriteChunk(chunk)
	assert.NoError(t, err)
	assert.Equal(t, len(paddedData)+internal.LengthChunkHeader, n)

	assert.NoError(t, writer.Close())

	actualContents := buffer.Bytes()
	assert.Equal(t, len(expectedStream), len(actualContents))
	assert.True(t, bytes.Equal(expectedStream, actualContents))
}

func expectationsNew(m *MockWriter) {
	var fileSize [4]byte // Empty bytes

	m.
		On("Write", goriffa.FourCCRIFF[:]).
		Return(len(goriffa.FourCCRIFF), test.NilError).
		Once()
	m.
		On("Write", fileSize[:]).
		Return(len(fileSize), test.NilError).
		Once()
	m.
		On("Write", test.FileType[:]).
		Return(len(test.FileType), test.NilError).
		Once()
}

func expectClose(m *MockWriter) {
	m.
		On("WriteAt", mock.Anything, int64(len(goriffa.FourCCRIFF))).
		Return(len(goriffa.FourCCRIFF)+4, test.NilError).
		Once()
}

type MockWriter struct {
	mock.Mock
}

var _ writer.WriterWithWriterAt = new(MockWriter)

func (m *MockWriter) Write(b []byte) (int, error) {
	args := m.Called(b)

	return args.Int(0), args.Error(1)
}

func (m *MockWriter) WriteAt(b []byte, offset int64) (int, error) {
	args := m.Called(b, offset)

	return args.Int(0), args.Error(1)
}
