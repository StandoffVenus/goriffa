# Goriffa

Goriffa is a package focused on writing and reading RIFF (Resource Interchange File Format) files. Wavefiles (.wav) and WEBP files are stored via RIFF, for example.

The main motivation for this library was (initially) to fork [go-riff](https://github.com/youpy/go-riff) and [go-wav](https://github.com/youpy/go-wav) in order to allow streaming/dynamic sizing of RIFF data. However, given the lack of updates for the aforementioned packages and the drastic API redesign, it made more sense for Goriffa to become its own repository.

Currently, Goriffa supports
- Reading RIFF data.
- Writing RIFF data and dynamically setting the data size RIFF field.
- Reading Wavefile format data (and samples - though, the sample will be raw bytes)
- Writing Wavefile data (including the format data)

# Okay, give me an example!

## Reading

Say you want to read a Wavefile - let's call it `cool-audio.wav`:

```golang
import (
    "os"

    "github.com/standoffvenus/goriffa/reader"
    "github.com/standoffvenus/goriffa/wave"
)

// Open the .wav file
f, fErr := os.Open("cool-audio.wav")
if fErr != nil {
    panic(fErr)
}
defer f.Close()

reader, readerErr := reader.New(f)
if readerErr != nil {
    panic(readerErr)
}

format, formatErr := wave.FormatFromReader(reader)
if formatErr != nil {
    panic(formatErr)
}

// Do something with the format data...
```

Then, assuming that the Wavefile contains raw PCM audio data, you could read the PCM data by reading from the initial reader:

```golang
var dataChunk goriffa.Chunk
if _, err := reader.ReadChunk(&dataChunk); err != nil {
    panic(err)
}

// Do something with the data chunk, which now contains
// the PCM data...
```

## Writing

Let's say we want to write a RIFF file. This is rather trivial with Goriffa:

```golang
import (
    "os"

    "github.com/standoffvenus/goriffa/writer"
)

// Open the .wav file
f, fErr := os.Open("cool-audio.wav")
if fErr != nil {
    panic(fErr)
}
defer f.Close()

writer, writerErr := writer.New(f)
if readerErr != nil {
    panic(readerErr)
}
defer writer.Close() // You must call Close when done with the writer or the dynamic content size is not persisted

someData := GetSomeData()
if _, err := writer.WriteChunk(goriffa.Chunk{
    Identifier: goriffa.FourCCData,
    Data: someData,
}); err != nil {
    panic(err)
}

// Hurray! We wrote the data!
```

# How do I execute/test locally?

To compile the application, immediately after cloning, generate the necessary Go code:

```
go generate ./...
```

Then, to test the project, you may run

```
go test ./...
```

The examplary WAVE and WEBP data used in some tests are embedded via `//go:embed`. Thanks, youpy, for packaging some example data with `go-riff`!