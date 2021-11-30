package internal_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/standoffvenus/goriffa/internal"
	"github.com/standoffvenus/goriffa/internal/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPaddedLengthNoPadding(t *testing.T) {
	assert.Equal(t, int64(4), internal.PaddedLength(4))
}

func TestPaddedLengthNeedsPadding(t *testing.T) {
	assert.Equal(t, int64(2), internal.PaddedLength(1))
}

func TestPadNoPadding(t *testing.T) {
	b := []byte{1, 2, 3, 4}
	assert.Equal(t, b, internal.Pad(b))
}

func TestPadNeedsPadding(t *testing.T) {
	b := []byte{1}
	assert.Equal(t, append(b, 0), internal.Pad(b))
}

func TestCopy(t *testing.T) {
	src := []byte{1, 2, 3, 4}
	dst := make([]byte, len(src))

	assert.NoError(t, internal.Copy(dst, src))
}

func TestCopyError(t *testing.T) {
	src := []byte{1, 2, 3, 4}
	dst := make([]byte, 0)

	assert.Error(t, internal.Copy(dst, src), io.ErrShortBuffer)
}

func TestRead(t *testing.T) {
	expectedBytes := [][]byte{{1, 2}, {3}, {4, 5}}
	expectedBuffer := bytes.NewReader(join(t, expectedBytes))
	length := expectedBuffer.Len()

	actualBytes := [][]byte{{0, 0}, {0}, {0, 0}}
	n, err := internal.Read(expectedBuffer, actualBytes...)
	assert.Equal(t, int64(length), n)
	assert.NoError(t, err)

	for idx, b := range actualBytes {
		assert.Equal(t, expectedBytes[idx], b)
	}
}

func TestReadError(t *testing.T) {
	err := errors.New("read error")
	reader := new(MockReadWriterAt)
	reader.
		On("Read", mock.Anything).
		Return(int(0), err).
		Once()

	n, e := internal.Read(reader, [][]byte{{}}...)
	assert.Equal(t, int64(0), n)
	assert.ErrorIs(t, e, err)

	reader.AssertExpectations(t)
	reader.AssertNumberOfCalls(t, "Read", 1)
}

func TestReadUnderflow(t *testing.T) {
	reader := new(MockReadWriterAt)
	reader.
		On("Read", mock.Anything).
		Return(int(0), test.NilError).
		Once()

	n, err := internal.Read(reader, [][]byte{{0}}...)
	assert.Equal(t, int64(0), n)
	assert.ErrorIs(t, err, internal.ErrBufferUnderflow)

	reader.AssertExpectations(t)
}

func TestWrite(t *testing.T) {
	expectedBytes := [][]byte{{1, 2}, {3}, {4, 5}}

	w := new(MockReadWriterAt)
	for _, b := range expectedBytes {
		w.
			On("Write", b).
			Return(len(b), test.NilError).
			Once()
	}

	n, err := internal.Write(w, expectedBytes...)
	assert.Equal(t, int64(len(join(t, expectedBytes))), n)
	assert.NoError(t, err)
}

func TestWriteError(t *testing.T) {
	err := errors.New("write error")
	w := new(MockReadWriterAt)
	w.
		On("Write", mock.Anything).
		Return(int(0), err).
		Once()

	n, e := internal.Write(w, [][]byte{{}}...)
	assert.Equal(t, int64(0), n)
	assert.ErrorIs(t, e, err)

	w.AssertExpectations(t)
	w.AssertNumberOfCalls(t, "Write", 1)
}

func TestWriteShortWrite(t *testing.T) {
	w := new(MockReadWriterAt)
	w.
		On("Write", mock.Anything).
		Return(int(0), test.NilError).
		Once()

	n, err := internal.Write(w, [][]byte{{0}}...)
	assert.Equal(t, int64(0), n)
	assert.ErrorIs(t, err, io.ErrShortWrite)

	w.AssertExpectations(t)
}

func TestWriteAt(t *testing.T) {
	expectedBytes := []byte{1, 2, 3, 4}
	writer := new(MockReadWriterAt)
	writer.
		On("WriteAt", expectedBytes, int64(1)).
		Return(len(expectedBytes), test.NilError).
		Once()

	n, err := internal.WriteAt(writer, expectedBytes, 1)
	assert.Equal(t, len(expectedBytes), n)
	assert.NoError(t, err)

	writer.AssertExpectations(t)
}

func TestWriteAtError(t *testing.T) {
	err := errors.New("read error")
	writer := new(MockReadWriterAt)
	writer.
		On("WriteAt", mock.Anything, int64(1)).
		Return(int(4), err).
		Once()

	n, e := internal.WriteAt(writer, []byte{}, 1)
	assert.Equal(t, int(4), n)
	assert.ErrorIs(t, e, err)

	writer.AssertExpectations(t)
}

func TestWriteAtUnderflow(t *testing.T) {
	writer := new(MockReadWriterAt)
	writer.
		On("WriteAt", mock.Anything, int64(1)).
		Return(int(0), test.NilError).
		Once()

	n, err := internal.WriteAt(writer, make([]byte, 1), 1)
	assert.Equal(t, int(0), n)
	assert.ErrorIs(t, err, io.ErrShortWrite)

	writer.AssertExpectations(t)
}

func join(t *testing.T, b [][]byte) []byte {
	var buffer bytes.Buffer
	for _, slice := range b {
		n, err := buffer.Write(slice)
		assert.Equal(t, n, len(slice))
		assert.NoError(t, err)
	}

	return buffer.Bytes()
}

type MockReadWriterAt struct {
	mock.Mock
}

func (m *MockReadWriterAt) Read(b []byte) (int, error) {
	args := m.Called(b)

	return args.Int(0), args.Error(1)
}

func (m *MockReadWriterAt) Write(b []byte) (int, error) {
	args := m.Called(b)

	return args.Int(0), args.Error(1)
}

func (m *MockReadWriterAt) WriteAt(b []byte, offset int64) (int, error) {
	args := m.Called(b, offset)

	return args.Int(0), args.Error(1)
}
