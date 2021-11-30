package reader

import (
	"bytes"
	"math"
	"testing"

	"github.com/standoffvenus/goriffa"
	"github.com/standoffvenus/goriffa/internal"
	"github.com/standoffvenus/goriffa/internal/test"
	"github.com/stretchr/testify/assert"
)

// Unfortunately, this test has to be in the reader package
// because there's no reasonable way to generate a near 4GB
// byte slice to cause overflows naturally. Instead, we'll
// just emulate the overflow scenario.

func TestReadOverflow(t *testing.T) {
	var buf bytes.Buffer
	if _, err := internal.Write(&buf,
		goriffa.FourCCRIFF[:],
		internal.LittleEndianUInt32Bytes(uint32(32)),
		test.FileType[:],
		[]byte{0, 14, 90, 32},
	); err != nil {
		panic(err)
	}

	r, err := New(&buf)
	assert.NoError(t, err)

	// Set the size super high.
	r.bytesRead = math.MaxUint32

	// Artificial overflow here.
	_, readErr := r.ReadChunk(new(internal.Chunk))
	assert.ErrorIs(t, readErr, goriffa.ErrCorrupted)
}
