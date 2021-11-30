package wave_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/standoffvenus/goriffa"
	"github.com/standoffvenus/goriffa/internal"
	"github.com/standoffvenus/goriffa/internal/test"
	"github.com/standoffvenus/goriffa/reader"
	"github.com/standoffvenus/goriffa/wave"
	"github.com/standoffvenus/goriffa/writer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var fmtExample = make([]byte, 0, 24)

func ExampleFormatFromReader() {
	// A Wavefile with the following properties
	// - PCM audio format
	// - Stereo channel
	// - 44100 Hz sampling rate
	// - 16-bit samples
	wavefile := SomeWavefile()
	reader, readerErr := reader.New(wavefile)
	if readerErr != nil {
		panic(readerErr)
	}

	format, formatErr := wave.FormatFromReader(reader)
	if formatErr != nil {
		panic(formatErr)
	}

	fmt.Printf("%dHz, %d-bit, %d-channel, %s audio.",
		format.SampleRate,
		format.BitsPerSample,
		format.Channels,
		format.AudioFormat)
	// Output: 44100Hz, 16-bit, 2-channel, PCM audio.
}

func ExampleWritePCM() {
	f := FileHandle()
	defer f.Close()

	w, wErr := writer.New(f, wave.FileTypeWavefile)
	if wErr != nil {
		panic(wErr)
	}
	defer w.Close()

	format := wave.Format{
		AudioFormat:   wave.PCM,
		SampleRate:    44100,
		Channels:      2,
		BitsPerSample: 16,
	}
	n, err := wave.WritePCM(w, format, SomePCMData())
	if err != nil {
		panic(err)
	}

	fmt.Printf("Wrote %d bytes.", n)
}

func FileHandle() *os.File {
	f, fErr := os.CreateTemp(os.TempDir(), "")
	if fErr != nil {
		panic(fErr)
	}

	return f
}

func SomePCMData() []byte {
	return []byte{42, 61, 79}
}

func SomeWavefile() io.Reader {
	data := make([]byte, 0)
	data = append(data, goriffa.FourCCRIFF[:]...)
	data = append(data, byte(len(wave.FileTypeWavefile))+byte(wave.LengthFormatChunk), 0, 0, 0)
	data = append(data, wave.FileTypeWavefile[:]...)
	data = append(data, goriffa.FourCCFormat[:]...)
	data = append(data, internal.LittleEndianUInt32Bytes(16)...)
	data = append(data, internal.LittleEndianUInt16Bytes(uint16(wave.PCM))...)
	data = append(data, internal.LittleEndianUInt16Bytes(2)...)
	data = append(data, internal.LittleEndianUInt32Bytes(44100)...)
	data = append(data, internal.LittleEndianUInt32Bytes(176400)...)
	data = append(data, internal.LittleEndianUInt16Bytes(4)...)
	data = append(data, internal.LittleEndianUInt16Bytes(16)...)

	return bytes.NewReader(data)
}

func TestMain(m *testing.M) {
	fmtExample = append(fmtExample, []byte("fmt ")...)
	fmtExample = append(fmtExample, internal.LittleEndianUInt32Bytes(16)...)
	fmtExample = append(fmtExample, internal.LittleEndianUInt16Bytes(uint16(wave.PCM))...)
	fmtExample = append(fmtExample, internal.LittleEndianUInt16Bytes(2)...)
	fmtExample = append(fmtExample, internal.LittleEndianUInt32Bytes(44100)...)
	fmtExample = append(fmtExample, internal.LittleEndianUInt32Bytes(176400)...)
	fmtExample = append(fmtExample, internal.LittleEndianUInt16Bytes(4)...)
	fmtExample = append(fmtExample, internal.LittleEndianUInt16Bytes(16)...)
	if len(fmtExample) != 24 {
		panic("format chunk became invalid size")
	}

	code := m.Run()

	os.Exit(code)
}

func TestNewFormat(t *testing.T) {
	f := wave.Format{
		AudioFormat:   wave.PCM,
		SampleRate:    44100,
		Channels:      2,
		BitsPerSample: 16,
	}

	assert.Equal(t, wave.PCM, f.AudioFormat)
	assert.Equal(t, uint32(44100), f.SampleRate)
	assert.Equal(t, uint16(2), f.Channels)
	assert.Equal(t, uint16(16), f.BitsPerSample)
}

func TestFormatFromReader(t *testing.T) {
	fmt, fmtErr := wave.FormatFromReader(exampleReader())
	assert.NoError(t, fmtErr)

	assert.Equal(t, wave.PCM, fmt.AudioFormat)
	assert.Equal(t, uint16(16), fmt.BitsPerSample)
	assert.Equal(t, uint16(2), fmt.Channels)
	assert.Equal(t, uint32(44100), fmt.SampleRate)
}

func TestFormatFromReaderWrongFOURCC(t *testing.T) {
	_, fmtErr := wave.FormatFromReader(readerFrom(&internal.Chunk{
		Identifier: goriffa.FourCCData,
		Data:       fmtExample[internal.LengthChunkHeader:],
	}))
	assert.ErrorIs(t, fmtErr, internal.ErrCorrupted)
}

func TestFormatFromReaderWrongBytesPerSecond(t *testing.T) {
	var badFmt [wave.LengthFormatChunk]byte
	copy(badFmt[:], fmtExample)
	badFmt[internal.LengthChunkHeader+8] = 0
	badFmt[internal.LengthChunkHeader+8+1] = 0
	badFmt[internal.LengthChunkHeader+8+2] = 0
	badFmt[internal.LengthChunkHeader+8+2] = 0

	_, fmtErr := wave.FormatFromReader(readerFrom(&internal.Chunk{
		Identifier: goriffa.FourCCFormat,
		Data:       badFmt[internal.LengthChunkHeader:],
	}))
	assert.ErrorIs(t, fmtErr, internal.ErrCorrupted)
}

func TestFormatFromReaderWrongBlockAlignment(t *testing.T) {
	var badFmt [wave.LengthFormatChunk]byte
	copy(badFmt[:], fmtExample)
	badFmt[internal.LengthChunkHeader+12] = 0
	badFmt[internal.LengthChunkHeader+12+1] = 0
	badFmt[internal.LengthChunkHeader+12+2] = 0
	badFmt[internal.LengthChunkHeader+12+2] = 0

	_, fmtErr := wave.FormatFromReader(readerFrom(&internal.Chunk{
		Identifier: goriffa.FourCCFormat,
		Data:       badFmt[internal.LengthChunkHeader:],
	}))
	assert.ErrorIs(t, fmtErr, internal.ErrCorrupted)
}

func TestFormatFromReaderError(t *testing.T) {
	err := errors.New("error")

	mockReader := new(MockGoriffaReadWriter)
	mockReader.
		On("ReadChunk", mock.Anything).
		Return(int(0), err).
		Once()

	_, fmtErr := wave.FormatFromReader(mockReader)
	assert.ErrorIs(t, fmtErr, err)
}

func TestFormatFromReaderWrongReadCount(t *testing.T) {
	mockReader := new(MockGoriffaReadWriter)
	mockReader.
		On("ReadChunk", mock.Anything).
		Return(int(wave.LengthFormatChunk+1), test.NilError).
		Once()
	mockReader.
		On("ReadChunk", mock.Anything).
		Return(int(wave.LengthFormatChunk-1), test.NilError).
		Once()

	_, fmtErrTooMany := wave.FormatFromReader(mockReader)
	_, fmtErrTooFew := wave.FormatFromReader(mockReader)

	assert.ErrorIs(t, fmtErrTooMany, goriffa.ErrCorrupted)
	assert.ErrorIs(t, fmtErrTooFew, goriffa.ErrCorrupted)
}

func TestWritePCM(t *testing.T) {
	pcm := []byte{4, 2}
	mockWriter := new(MockGoriffaReadWriter)
	mockWriter.
		On("WriteChunk", goriffa.Chunk{
			Identifier: goriffa.FourCCFormat,
			Data:       fmtExample[internal.LengthChunkHeader:],
		}).
		Return(len(fmtExample), test.NilError).
		Once()
	mockWriter.
		On("WriteChunk", goriffa.Chunk{
			Identifier: goriffa.FourCCData,
			Data:       pcm,
		}).
		Return(int(internal.LengthChunkHeader)+len(pcm), test.NilError).
		Once()

	fmt, fmtErr := wave.FormatFromReader(exampleReader())
	assert.NoError(t, fmtErr)

	writeN, writeErr := wave.WritePCM(mockWriter, fmt, pcm)
	assert.NoError(t, writeErr)
	assert.Equal(t, len(fmtExample)+internal.LengthChunkHeader+len(pcm), writeN)

	mockWriter.AssertExpectations(t)
}

func TestWritePCMFormatError(t *testing.T) {
	err := errors.New("error")
	mockWriter := new(MockGoriffaReadWriter)
	mockWriter.
		On("WriteChunk", goriffa.Chunk{
			Identifier: goriffa.FourCCFormat,
			Data:       fmtExample[internal.LengthChunkHeader:],
		}).
		Return(len(fmtExample), err).
		Once()

	fmt, fmtErr := wave.FormatFromReader(exampleReader())
	assert.NoError(t, fmtErr)

	writeN, writeErr := wave.WritePCM(mockWriter, fmt, nil)
	assert.ErrorIs(t, writeErr, err)
	assert.Equal(t, len(fmtExample), writeN)

	mockWriter.AssertExpectations(t)
}

func TestWritePCMDataError(t *testing.T) {
	err := errors.New("error")
	pcm := []byte{4, 2}
	mockWriter := new(MockGoriffaReadWriter)
	mockWriter.
		On("WriteChunk", goriffa.Chunk{
			Identifier: goriffa.FourCCFormat,
			Data:       fmtExample[internal.LengthChunkHeader:],
		}).
		Return(len(fmtExample), test.NilError).
		Once()
	mockWriter.
		On("WriteChunk", goriffa.Chunk{
			Identifier: goriffa.FourCCData,
			Data:       pcm,
		}).
		Return(int(internal.LengthChunkHeader)+len(pcm), err).
		Once()

	fmt, fmtErr := wave.FormatFromReader(exampleReader())
	assert.NoError(t, fmtErr)

	writeN, writeErr := wave.WritePCM(mockWriter, fmt, pcm)
	assert.ErrorIs(t, writeErr, err)
	assert.Equal(t, len(fmtExample)+internal.LengthChunkHeader+len(pcm), writeN)

	mockWriter.AssertExpectations(t)
}

func TestReadWavefile(t *testing.T) {
	r, details := test.WAV()
	reader, newErr := reader.New(r)
	assert.NoError(t, newErr)

	fmt, fmtErr := wave.FormatFromReader(reader)
	assert.NoError(t, fmtErr)

	assert.Equal(t, details.SampleRate, fmt.SampleRate)
	assert.Equal(t, details.Channels, fmt.Channels)
	assert.Equal(t, wave.AudioFormat(details.AudioFormat), fmt.AudioFormat)
	assert.Equal(t, details.BitsPerSample, fmt.BitsPerSample)
}

func exampleReader() goriffa.Reader {
	return readerFrom(&internal.Chunk{
		Identifier: goriffa.FourCCFormat,
		Data:       fmtExample[internal.LengthChunkHeader:],
	})
}

func readerFrom(ch *internal.Chunk) goriffa.Reader {
	mockReader := new(MockGoriffaReadWriter)
	mockReader.
		On("ReadChunk", mock.Anything).
		Return(len(fmtExample), test.NilError).
		Run(func(args mock.Arguments) {
			c := args.Get(0).(*internal.Chunk)
			*c = *ch
		}).
		Once()

	return mockReader
}

type MockGoriffaReadWriter struct {
	mock.Mock
}

func (m *MockGoriffaReadWriter) ReadChunk(ch *internal.Chunk) (int, error) {
	args := m.Called(ch)

	return args.Int(0), args.Error(1)
}

func (m *MockGoriffaReadWriter) WriteChunk(ch internal.Chunk) (int, error) {
	args := m.Called(ch)

	return args.Int(0), args.Error(1)
}
