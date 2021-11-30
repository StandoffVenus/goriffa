package writer_test

import (
	"fmt"
	"os"

	"github.com/standoffvenus/goriffa"
	"github.com/standoffvenus/goriffa/internal"
	"github.com/standoffvenus/goriffa/wave"
	"github.com/standoffvenus/goriffa/writer"
)

func Example() {
	f, close, err := OpenTemp()
	if err != nil {
		panic(err)
	}
	defer close()

	riffWriter, writerErr := writer.New(f, wave.FileTypeWavefile)
	if writerErr != nil {
		panic(writerErr)
	}
	defer riffWriter.Close()

	n, err := riffWriter.WriteChunk(internal.Chunk{
		Identifier: goriffa.FourCCData,
		Data:       []byte("Hello, world!"),
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Wrote %d bytes.", n)

	// Output: Wrote 22 bytes.
}

func OpenTemp() (*os.File, func() error, error) {
	f, err := os.CreateTemp(os.TempDir(), "")
	if err != nil {
		return nil, nil, err
	}

	return f, f.Close, nil
}
