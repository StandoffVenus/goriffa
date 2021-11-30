package reader_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/standoffvenus/goriffa"
	"github.com/standoffvenus/goriffa/internal"
	"github.com/standoffvenus/goriffa/internal/test"
	"github.com/standoffvenus/goriffa/reader"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func ExampleReader_ReadChunk() {
	r, rErr := reader.New(RIFFReader())
	if rErr != nil {
		panic(rErr)
	}

	var chunk internal.Chunk
	if n, err := r.ReadChunk(&chunk); err != nil {
		panic(err)
	} else {
		fmt.Printf("Read %d bytes.\n", n)
	}

	PrintChunk(chunk)
	// Output: Read 14 bytes.
	// Chunk { ID: "fmt ", Size: 5, Data: [1 2 3 4 5] }
}

func TestNew(t *testing.T) {
	r, err := reader.New(bytes.NewReader(header(0)))
	assert.NoError(t, err)

	assert.Equal(t, test.FileType, r.FileType())
	assert.Equal(t, uint32(4), r.Size())
}

func TestNewReadError(t *testing.T) {
	expectedError := errors.New("error")
	mockReader := new(MockReader)
	mockReader.PrepareRead([]byte{}, expectedError)

	_, err := reader.New(mockReader)
	assert.ErrorIs(t, err, expectedError)

	mockReader.AssertExpectations(t)
}

func TestNewReadUnderflowError(t *testing.T) {
	mockReader := new(MockReader)
	mockReader.PrepareRead([]byte{}, internal.ErrBufferUnderflow)

	_, err := reader.New(mockReader)
	assert.ErrorIs(t, err, goriffa.ErrCorrupted)

	mockReader.AssertExpectations(t)
}

func TestNewNoRIFFHeader(t *testing.T) {
	header := header(0)
	header = header[4:]
	var emptyBytes [4]byte

	_, err := reader.New(bytes.NewReader(append(emptyBytes[:], header...)))
	assert.ErrorIs(t, err, goriffa.ErrCorrupted)
}

func TestReadChunk(t *testing.T) {
	expectedChunk := goriffa.Chunk{
		Identifier: goriffa.FourCCData,
		Data:       []byte{1, 2, 3, 4},
	}

	var buf bytes.Buffer
	buf.Write(header(expectedChunk.ByteLength()))
	buf.Write(chunk(expectedChunk))

	r, err := reader.New(&buf)
	assert.NoError(t, err)

	var chunk goriffa.Chunk
	n, readErr := r.ReadChunk(&chunk)
	assert.NoError(t, readErr)
	assert.Equal(t, expectedChunk.ByteLength(), int64(n))

	assert.Equal(t, expectedChunk.Identifier, chunk.Identifier)
	assert.Equal(t, expectedChunk.Data, expectedChunk.Data)
}

func TestReadPaddedChunk(t *testing.T) {
	expectedChunk := goriffa.Chunk{
		Identifier: goriffa.FourCCData,
		Data:       []byte{1, 2, 3, 4, 5},
	}

	var buf bytes.Buffer
	buf.Write(header(expectedChunk.ByteLength()))
	buf.Write(chunk(expectedChunk))

	r, err := reader.New(&buf)
	assert.NoError(t, err)

	var chunk goriffa.Chunk
	n, readErr := r.ReadChunk(&chunk)
	assert.NoError(t, readErr)
	assert.Equal(t, internal.PaddedLength(expectedChunk.ByteLength()), int64(n))

	assert.Equal(t, expectedChunk.Identifier, chunk.Identifier)
	assert.Equal(t, expectedChunk.Data, expectedChunk.Data)
}

func TestReadChunkErrorOnFirstRead(t *testing.T) {
	var err error = errors.New("read error")
	mockReader := new(MockReader)
	expectNew(mockReader, 32)
	mockReader.PrepareRead([]byte{}, err)

	r, newErr := reader.New(mockReader)
	assert.NoError(t, newErr)

	var chunk goriffa.Chunk
	_, readErr := r.ReadChunk(&chunk)
	assert.ErrorIs(t, readErr, err)

	mockReader.AssertExpectations(t)
}

func TestReadHeaderUnderflow(t *testing.T) {
	mockReader := new(MockReader)
	expectNew(mockReader, 32)
	mockReader.PrepareRead([]byte{4, 3, 1}, nil)

	r, newErr := reader.New(mockReader)
	assert.NoError(t, newErr)

	var chunk goriffa.Chunk
	_, readErr := r.ReadChunk(&chunk)
	assert.ErrorIs(t, readErr, goriffa.ErrCorrupted)

	mockReader.AssertExpectations(t)
}

func TestReadChunkErrorOnSecondRead(t *testing.T) {
	var err error = errors.New("read error")
	mockReader := new(MockReader)
	expectNew(mockReader, 32)
	mockReader.PrepareRead(make([]byte, 8), nil)
	mockReader.PrepareRead([]byte{}, err)

	r, newErr := reader.New(mockReader)
	assert.NoError(t, newErr)

	var chunk goriffa.Chunk
	n, readErr := r.ReadChunk(&chunk)
	assert.Equal(t, n, int(8))
	assert.ErrorIs(t, readErr, err)

	mockReader.AssertExpectations(t)
}

func TestReadChunkUnderflowOnData(t *testing.T) {
	mockReader := new(MockReader)
	expectNew(mockReader, 32)
	mockReader.PrepareRead([]byte{42, 42, 42, 42, 42, 0, 0, 0}, nil)
	mockReader.PrepareRead([]byte{}, nil)

	r, newErr := reader.New(mockReader)
	assert.NoError(t, newErr)

	var chunk goriffa.Chunk
	n, readErr := r.ReadChunk(&chunk)
	assert.Equal(t, n, int(8))
	assert.ErrorIs(t, readErr, goriffa.ErrCorrupted)

	mockReader.AssertExpectations(t)
}

func header(size int64) []byte {
	headerBytes := make([]byte, 0, 12)
	buf := bytes.NewBuffer(headerBytes[:])

	if _, err := internal.Write(buf,
		goriffa.FourCCRIFF[:],
		internal.LittleEndianUInt32Bytes(uint32(size+4)),
		test.FileType[:],
	); err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func chunk(c internal.Chunk) []byte {
	return append(
		c.Identifier[:],
		append(
			internal.LittleEndianUInt32Bytes(uint32(len(c.Data))),
			internal.Pad(c.Data)...)...)
}

func expectNew(r *MockReader, fileSize uint32) {
	r.PrepareRead(goriffa.FourCCRIFF[:], nil)
	r.PrepareRead(internal.LittleEndianUInt32Bytes(fileSize), nil)
	r.PrepareRead(test.FileType[:], nil)
}

type MockReader struct {
	mock.Mock
}

var _ io.Reader = new(MockReader)

func (m *MockReader) PrepareRead(b []byte, err error) {
	m.
		On("Read", mock.AnythingOfType(fmt.Sprintf("%T", []byte{}))).
		Return(len(b), err).
		Run(func(args mock.Arguments) {
			copy(args.Get(0).([]byte), b)
		}).
		Once()
}

func (m *MockReader) Read(b []byte) (int, error) {
	args := m.Called(b)

	return args.Int(0), args.Error(1)
}
