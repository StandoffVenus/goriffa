package writer

import (
	"math"
	"testing"

	"github.com/standoffvenus/goriffa"
	"github.com/standoffvenus/goriffa/internal"
	"github.com/standoffvenus/goriffa/internal/test"
	"github.com/stretchr/testify/assert"
)

// Unfortunately, this test has to be in the writer package
// because there's on reasonable way to generate enough
// traffic to the Write method to overflow it (we're talking
// like 4 GB's of data). Instead, we'll just emulate the
// overflow scenario.

func TestWriteOverflow(t *testing.T) {
	var buf test.Buffer
	w, err := New(&buf, test.FileType)
	defer func() { _ = w.Close() }()
	assert.NoError(t, err)

	// Set the file size super high.
	w.fileSize = math.MaxUint32

	// Artificial overflow here.
	n, err := w.WriteChunk(internal.Chunk{})
	assert.Equal(t, int(0), n)
	assert.ErrorIs(t, err, goriffa.ErrCorrupted)
}
