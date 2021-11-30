// Package goriffa is a package focused on writing and
// reading RIFF (Resource Interchange File Format) files.
// Wavefiles (.wav) and WEBP files are stored via RIFF,
// for example.
//
// A RIFF file consists of two components:
//
// - The 12-byte RIFF header.
//
// - The remaining chunks.
//
// The RIFF header always starts a RIFF file. The
// header must start with the ASCII characters "RIFF",
// followed by 4-bytes holding the total length of the
// data succeeding it. The last 4-bytes of the header
// contain the file type (e.g. "WAVE" for a Wavefile)
//
// The remaining chunks are formatted as below:
//  <FOURCC>, <chunk length>, <chunk>
// where "FOURCC" is a 4-byte chunk identifier,
// "chunk length" holds the length of the preceding
// "chunk" - the actual data of the chunk.
//
// For more information, see Microsoft's overview
// of RIFF: https://docs.microsoft.com/en-us/windows/win32/xaudio2/resource-interchange-file-format--riff-.
package goriffa

import "github.com/standoffvenus/goriffa/internal"

// FileType represents a RIFF file type, like
// "WAVE" or "WEBP".
type FileType = internal.FileType

// Chunk represents a RIFF chunk.
type Chunk = internal.Chunk

type (
	// Reader represents a RIFF data reader.
	Reader interface {
		// ReadChunk reads from the underlying
		// reader and parses the data into the
		// provided chunk pointer, returning how
		// many bytes were read from the underlying
		// reader and an error if one occurred.
		ReadChunk(*Chunk) (int, error)
	}

	// Writer represents a RIFF data writer.
	Writer interface {
		// WriteChunk writes the provided chunk,
		// returning how many bytes were written
		// to the underlying writer and an error
		// if one occurred.
		WriteChunk(Chunk) (int, error)
	}
)

// Valid FOURCC identifiers.
//
// The type of data in a chunk is indicated by a four-character code
// (FOURCC) identifier. A FOURCC is a 32-bit unsigned integer created
// by concatenating four ASCII characters used to identify chunk types
// in a RIFF file.
var (
	FourCCRIFF   internal.FourCC = internal.FourCC(internal.StringMust4Byte("RIFF"))
	FourCCFormat internal.FourCC = internal.FourCC(internal.StringMust4Byte("fmt "))
	FourCCData   internal.FourCC = internal.FourCC(internal.StringMust4Byte("data"))
	FourCCSMPL   internal.FourCC = internal.FourCC(internal.StringMust4Byte("smpl"))
	FourCCWSMP   internal.FourCC = internal.FourCC(internal.StringMust4Byte("wsmp"))
)
