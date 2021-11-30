package reader_test

import (
	"bytes"
	"fmt"
	"io"

	"github.com/standoffvenus/goriffa"
	"github.com/standoffvenus/goriffa/reader"
	"github.com/standoffvenus/goriffa/wave"
)

func Example() {
	riffReader, readerErr := reader.New(RIFFReader())
	if readerErr != nil {
		panic(readerErr)
	}

	chunks, readAllErr := riffReader.ReadToEnd()
	if readAllErr != nil {
		panic(readAllErr)
	}

	for _, c := range chunks {
		PrintChunk(c)
	}

	// Output: Chunk { ID: "fmt ", Size: 5, Data: [1 2 3 4 5] }
	// Chunk { ID: "data", Size: 8, Data: [1 2 3 4 5 6 7 8] }
}

func PrintChunk(ch goriffa.Chunk) {
	fmt.Printf("Chunk { ID: %q, Size: %d, Data: %v }\n",
		ch.Identifier,
		len(ch.Data),
		ch.Data)
}

func RIFFReader() io.Reader {
	// RIFF files have 12 byte headers.
	// RIFF files are also little-endian.
	//
	// The data below is equivalent to
	// {RIFF 42 WAVE}
	// i.e. a Wavefile that is 42 bytes long.

	riffData := make([]byte, 0, 12)
	riffData = append(riffData, goriffa.FourCCRIFF[:]...)
	riffData = append(riffData, 42, 0, 0, 0)
	riffData = append(riffData, wave.FileTypeWavefile[:]...)

	// We define two chunks below:
	// 1. 5-byte chunk, identified as "fmt " (it must
	//    be padded, as demanded by the RIFF spec)
	// 2. 8-byte chunk, identified as "data"
	riffData = append(riffData, goriffa.FourCCFormat[:]...)
	riffData = append(riffData, 5, 0, 0, 0)
	riffData = append(riffData, 1, 2, 3, 4, 5, 0)
	riffData = append(riffData, goriffa.FourCCData[:]...)
	riffData = append(riffData, 8, 0, 0, 0)
	riffData = append(riffData, 1, 2, 3, 4, 5, 6, 7, 8)

	return bytes.NewReader(riffData)
}
