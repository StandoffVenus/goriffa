package internal

// LengthChunkHeader represents the length (in bytes) of
// a RIFF chunk header.
const LengthChunkHeader int = 8

// EmptyBytes holds an empty 4 bytes.
var EmptyBytes fourBytes

type fourBytes = [4]byte

type (
	// FourCC represents the RIFF FOURCC bytes.
	FourCC fourBytes

	// FileType represents the RIFF content file type.
	FileType fourBytes
)

// Chunk represents a RIFF chunk.
type Chunk struct {
	// The chunk's FOURCC identifier.
	Identifier FourCC

	// Data holds the chunk data. The
	// chunk's size is inferred from
	// the length of Data.
	Data []byte
}

// ByteLength will return how many bytes this
// chunk would be once formatted (including
// header and padding)
func (c Chunk) ByteLength() int64 {
	return 8 + PaddedLength(int64(len(c.Data)))
}

// String will return the string representation
// of the FOURCC.
func (cc FourCC) String() string {
	return string(cc[:])
}

// string will return the string representation
// of the file type.
func (ft FileType) String() string {
	return string(ft[:])
}
