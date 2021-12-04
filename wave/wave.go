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

var (
	// ErrBadWaveData occurs when errors in Wavefile
	// data are recognized, such as invalid block
	// alignment, unrecognized audio formats, etc.
	//
	// ErrBadWaveData wraps goriffa.ErrCorrupted so
	//  if errors.Is(err, goriffa.ErrCorrupted) {
	// can be used to generally capture errors when
	// reading/writing Wavefile data.
	//
	// All Wavefile errors wrap ErrBadWaveData.
	ErrBadWaveData error = internal.Wrap("invalid Wavefile data", internal.ErrCorrupted)

	// ErrUnrecognizedAudioFormat occurs when an unrecognized
	// audio format integer is in the Wavefile data.
	ErrUnrecognizedAudioFormat error = internal.Wrap("unrecognized audio format", ErrBadWaveData)

	// ErrInvalidChannelCount occurs when a Wavefile's
	// channel count is 0.
	ErrInvalidChannelCount error = internal.Wrap("invalid channel count", ErrBadWaveData)

	// ErrInvalidSampleRate occurs when a Wavefile's
	// sample rate is 0.
	ErrInvalidSampleRate error = internal.Wrap("invalid sample rate", ErrBadWaveData)

	// ErrInconsistentBytesPerSecond occurs when the average
	// bytes-per-second field of the Wavefile data
	// does not match SampleRate * Channels * BitRate / 8.
	ErrInconsistentBytesPerSecond error = internal.Wrap("bytes per second was inconsistent", ErrBadWaveData)

	// ErrInconsistentBlockAlignment occurs when the block
	// alignment bytes for a Wavefile is erroneous
	// given the channel count and bits per sample.
	ErrInconsistentBlockAlignment error = internal.Wrap("block alignment was inconsistent", ErrBadWaveData)

	// ErrInvalidBitRate occurs when a Wavefile's bit rate
	// is 0 or not a multiple of 8.
	ErrInvalidBitRate error = internal.Wrap("invalid bit rate", ErrBadWaveData)
)

// FileTypeWavefile represents the file type
// stored in a Wavefile ("WAVE")
var FileTypeWavefile internal.FileType = internal.Must4Byte([]byte("WAVE"))

// WaveFormat represents the format of
// the audio for the Wavefile, like PCM.
type AudioFormat uint16

// Format represents details about a
// Wavefile.
//
// The format chunk for Wavefile data is
// formatted as follows (all numbers are
// little-endian):
//  [24]byte{
//    fmt                    // 4 bytes for chunk identifier ("fmt " in this case)
//    0x10, 0, 0, 0          // 4 bytes for chunk length (always 16 for format chunks)
//    1, 0                   // 2 bytes for audio format (1 represents PCM)
//    2, 0                   // 2 bytes for channel count (2 channels)
//    0x44, 0xAC, 0, 0       // 4 bytes for sample rate (44100Hz)
//    0x10, 0xB1, 0x02, 0x00 // 4 bytes expected to be (Sample Rate * BitsPerSample * Channels) / 8 == 172400
//    4, 0                   // 2 bytes for block alignment: (BitsPerSample / 8 * Channels) == 4
//    0x10, 0                // 2 bytes for bits per sample
//  }
type Format struct {
	// BitsPerSample will return how many
	// bits a sample in the Wavefile is
	// encoded with.
	// Bits per sample must be a non-zero
	// multiple of 8.
	BitsPerSample uint16

	// Channels will return the number of
	// channels of a Wavefile.
	// While the channel count cannot be
	// 0, any other channel count is
	// technically allowed.
	Channels uint16

	// SampleRate will return the sample
	// rate of a Wavefile.
	// Sample rate
	SampleRate uint32

	// AudioFormat will return the audio
	// format of the Wavefile, such as
	// PCM.
	// The audio format must be a supported
	// audio format.
	//
	// TODO: Support more than PCM.
	AudioFormat AudioFormat
}

// WritePCM will write PCM data to the provided RIFF writer
// in the Wavefile format. The number of bytes written and
// any error will be returned. If the Wavefile format fails
// to validate, the corresponding error is returned. If the
// provided format's audio format is not PCM, ErrBadWaveData
// is returned.
//
// The provided format will be written as the Wavefile's format
// header, whereas the PCM data provided will be written in the
// Wavefile's data section.
//
// The writer is expected to be initialized - that is, the RIFF
// header is expected to be written to the writer already.
func WritePCM(w goriffa.Writer, f Format, pcm []byte) (int, error) {
	if err := Validate(f); err != nil {
		return 0, err
	}
	if f.AudioFormat != PCM {
		return 0, fmt.Errorf(
			"%w: attempt to write PCM data with non-PCM audio format",
			ErrBadWaveData)
	}

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
		return f, fmt.Errorf("%w: format chunk is invalid size (%d)", ErrBadWaveData, n)
	}

	if ch.Identifier != goriffa.FourCCFormat {
		return f, fmt.Errorf("%w: format chunk FOURCC incorrect (%s, should be %s)",
			ErrBadWaveData,
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
			ErrInconsistentBytesPerSecond,
			expectedBytesPerSecond,
			f.BytesPerSecond(),
		)
	}

	if expectedBlockAlign != uint16(f.BlockAlign()) {
		return f, fmt.Errorf(
			"%w: stream contained invalid block alignment (expected %d, was %d)",
			ErrInconsistentBlockAlignment,
			expectedBlockAlign,
			f.BlockAlign(),
		)
	}

	if err := Validate(f); err != nil {
		return f, err
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
		internal.Panic(err)
	}

	return rawBytes
}

// Validate will return an error if the provided
// format is invalid.
func Validate(f Format) error {
	switch {
	case f.BitsPerSample == 0:
		return fmt.Errorf(
			"%w: bit rate cannot be 0",
			ErrInvalidBitRate,
		)
	case f.BitsPerSample%8 != 0:
		return fmt.Errorf(
			"%w: bit rate must be a multiple of 8 (was %d)",
			ErrInvalidBitRate,
			f.BitsPerSample,
		)
	}

	if f.Channels == 0 {
		return fmt.Errorf(
			"%w: channel count cannot be 0",
			ErrInvalidChannelCount)
	}

	if f.SampleRate == 0 {
		return fmt.Errorf(
			"%w: sample rate cannot be 0",
			ErrInvalidSampleRate)
	}

	return nil
}
