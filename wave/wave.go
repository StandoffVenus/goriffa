//go:generate stringer -type=AudioFormat

// Package wave defines mechanisms for reading, writing,
// and understanding Wavefile data. The primary use of
// package comes with its Format type; the Format type
// contains functionality that simplifies reading and
// writing Wavefile format data (e.g., sampling
// rate, bits-per-sample, etc.)
package wave

import (
	"bytes"
	"fmt"

	"github.com/standoffvenus/goriffa"
	"github.com/standoffvenus/goriffa/internal"
)

// LengthFormatChunk is the length of the format
// chunk every Wavefile has.
const LengthFormatChunk int = internal.LengthChunkHeader + 16

// Recognized Wavefile audio formats.
const (
	PCM AudioFormat = 1
)

// FileTypeWavefile represents the file type
// stored in a Wavefile ("WAVE")
var FileTypeWavefile internal.FileType = internal.Must4Byte([]byte("WAVE"))

// WaveFormat represents the format of
// the audio for the Wavefile, like PCM.
type AudioFormat uint16

// Format represents details about a
// Wavefile.
type Format struct {
	// BitsPerSample will return how many
	// bits a sample in the Wavefile is
	// encoded with.
	BitsPerSample uint16

	// Channels will return the number of
	// channels of a Wavefile.
	Channels uint16

	// SampleRate will return the sample
	// rate of a Wavefile.
	SampleRate uint32

	// AudioFormat will return the audio
	// format of the Wavefile, such as
	// PCM.
	AudioFormat AudioFormat
}

// WritePCM will write PCM data to the provided RIFF writer
// in the Wavefile format. The number of bytes written and
// any error will be returned.
//
// The provided format will be written as the Wavefile's format
// header, whereas the PCM data provided will be written in the
// Wavefile's data section.
//
// The writer is expected to be initialized - that is, the RIFF
// header is expected to be written to the writer already.
func WritePCM(w goriffa.Writer, f Format, pcm []byte) (int, error) {
	data := formatToBytes(f)
	fmtN, fmtErr := w.WriteChunk(internal.Chunk{
		Identifier: goriffa.FourCCFormat,
		Data:       data[internal.LengthChunkHeader:],
	})
	if fmtErr != nil {
		return fmtN, fmtErr
	}

	audioN, audioErr := w.WriteChunk(internal.Chunk{
		Identifier: goriffa.FourCCData,
		Data:       pcm,
	})

	return fmtN + audioN, audioErr
}

// FormatFromReader will parse the Wavefile format header
// from the provided reader. If the format is invalid, an
// error is returned.
//
// FormatFromReader reads a single chunk from the reader,
// expecting it to be the format of the Wavefile.
func FormatFromReader(r goriffa.Reader) (Format, error) {
	var f Format

	var ch internal.Chunk
	if n, err := r.ReadChunk(&ch); err != nil {
		return f, err
	} else if n != LengthFormatChunk {
		return f, fmt.Errorf("%w: format chunk is invalid size (%d)", internal.ErrCorrupted, n)
	}

	if ch.Identifier != goriffa.FourCCFormat {
		return f, fmt.Errorf("%w: format chunk FOURCC incorrect (%s, should be %s)",
			internal.ErrCorrupted,
			string(ch.Identifier[:]),
			string(goriffa.FourCCFormat[:]))
	}

	// Once we've reached this point, we know that the
	// chunk's data should be the valid length to do
	// the below operations; ergo, we don't need to worry
	// about panics.
	buffer := bytes.NewReader(ch.Data)
	f.AudioFormat = AudioFormat(internal.ReadLittleEndianUInt16(buffer))
	f.Channels = internal.ReadLittleEndianUInt16(buffer)
	f.SampleRate = internal.ReadLittleEndianUInt32(buffer)
	expectedBytesPerSecond := internal.ReadLittleEndianUInt32(buffer)
	expectedBlockAlign := internal.ReadLittleEndianUInt16(buffer)
	f.BitsPerSample = internal.ReadLittleEndianUInt16(buffer)

	if expectedBytesPerSecond != uint32(f.BytesPerSecond()) {
		return f, fmt.Errorf(
			"%w: stream has invalid average bytes-per-second field (expected %d, was %d)",
			internal.ErrCorrupted,
			expectedBytesPerSecond,
			f.BytesPerSecond(),
		)
	}

	if expectedBlockAlign != uint16(f.BlockAlign()) {
		return f, fmt.Errorf(
			"%w: stream contained invalid block alignment (expected %d, was %d)",
			internal.ErrCorrupted,
			expectedBlockAlign,
			f.BlockAlign(),
		)
	}

	return f, nil
}

// BlockAlign returns the block alignment,
// i.e.:
//  f.BitsPerSample / 8 * f.Channels
func (f Format) BlockAlign() uint16 {
	return f.BitsPerSample / 8 * f.Channels
}

// BytesPerSecond returns the bytes per second
// of audio, i.e.:
//  f.SampleRate * f.BitsPerSample / 8 * f.Channels
func (f Format) BytesPerSecond() uint32 {
	return f.SampleRate * uint32(f.BitsPerSample) / 8 * uint32(f.Channels)
}

func formatToBytes(f Format) [LengthFormatChunk]byte {
	formatData := make([]byte, 0, LengthFormatChunk-internal.LengthChunkHeader)
	formatData = append(formatData, internal.LittleEndianUInt16Bytes(uint16(f.AudioFormat))...)
	formatData = append(formatData, internal.LittleEndianUInt16Bytes(f.Channels)...)
	formatData = append(formatData, internal.LittleEndianUInt32Bytes(f.SampleRate)...)
	formatData = append(formatData, internal.LittleEndianUInt32Bytes(uint32(f.BytesPerSecond()))...)
	formatData = append(formatData, internal.LittleEndianUInt16Bytes(uint16(f.BlockAlign()))...)
	formatData = append(formatData, internal.LittleEndianUInt16Bytes(uint16(f.BitsPerSample))...)

	formatBuffer := make([]byte, 0, LengthFormatChunk)
	formatBuffer = append(formatBuffer, goriffa.FourCCFormat[:]...)
	formatBuffer = append(formatBuffer, internal.LittleEndianUInt32Bytes(uint32(len(formatData)))...)
	formatBuffer = append(formatBuffer, formatData...)

	var rawBytes [LengthFormatChunk]byte
	if err := internal.Copy(rawBytes[:], formatBuffer); err != nil {
		panic(err)
	}

	return rawBytes
}
