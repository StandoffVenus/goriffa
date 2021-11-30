package test

import (
	"bytes"
	"io"
	"testing"

	_ "embed"

	"github.com/standoffvenus/goriffa/internal"
	"github.com/standoffvenus/goriffa/wave"
)

// Long will skip the test it's called from if
// short testing is enabled.
func Long(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
}

// NilError may be used whenever mocking to provide
// a nil value of type error - avoids type assertion
// errors.
var NilError error = nil

// FileType is a simple file type constant to
// be used for tests.
var FileType internal.FileType = internal.FileType{10, 12, 14, 16}

//go:embed files/test.wav
var wav []byte

//go:embed files/test.webp
var webp []byte

// Details holds details about a
// given RIFF test file.
type Details interface {
	FileType() internal.FileType
	Size() uint32
}

type details struct {
	fileType internal.FileType
	size     uint32
}

// WavefileDetails holds details
// specifically about a Wavefile.
type WavefileDetails struct {
	details
	wave.Format
}

// WebPDetails holds details
// specifically about a WEBP file.
type WebPDetails struct {
	details
}

// WAV will return a reader that
// returns the contents of the .wav
// test file.
func WAV() (io.Reader, WavefileDetails) {
	return bytes.NewReader(wav), WavefileDetails{
		details: details{
			size:     243800,
			fileType: internal.Must4Byte([]byte("WAVE")),
		},
		Format: wave.Format{
			SampleRate:    44100,
			Channels:      2,
			AudioFormat:   wave.PCM,
			BitsPerSample: 16,
		},
	}
}

// WEBP will return a reader that
// returns the contents of the .webp
// test file.
func WEBP() (io.Reader, WebPDetails) {
	return bytes.NewReader(webp), WebPDetails{
		details: details{
			size:     23396,
			fileType: internal.Must4Byte([]byte("WEBP")),
		},
	}
}

func (d details) FileType() internal.FileType {
	return d.fileType
}

func (d details) Size() uint32 {
	return d.size
}
