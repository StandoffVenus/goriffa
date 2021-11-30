package wave_test

import (
	"fmt"
	"io"

	"github.com/standoffvenus/goriffa/internal"
	"github.com/standoffvenus/goriffa/internal/test"
	"github.com/standoffvenus/goriffa/reader"
	"github.com/standoffvenus/goriffa/wave"
)

func Example() {
	reader, readerErr := reader.New(Wavefile())
	if readerErr != nil {
		panic(readerErr)
	}

	format, formatErr := wave.FormatFromReader(reader)
	if formatErr != nil {
		panic(formatErr)
	}

	fmt.Println("Format:")
	fmt.Printf("  Sample rate: %dHz | Channels: %d | Bits Per Sample: %d | Audio Format: %s\n",
		format.SampleRate,
		format.Channels,
		format.BitsPerSample,
		format.AudioFormat)

	// Read the rest of the Wavefile data, i.e. the PCM
	// data in this case.
	var ch internal.Chunk
	if _, err := reader.ReadChunk(&ch); err != nil {
		panic(err)
	}

	// Output: Format:
	//   Sample rate: 44100Hz | Channels: 2 | Bits Per Sample: 16 | Audio Format: PCM
}

func Wavefile() io.Reader {
	r, _ := test.WAV()

	return r
}
