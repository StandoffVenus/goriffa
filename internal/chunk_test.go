package internal_test

import (
	"testing"

	"github.com/standoffvenus/goriffa"
	"github.com/standoffvenus/goriffa/internal"
	"github.com/stretchr/testify/assert"
)

func TestByteLength(t *testing.T) {
	c := internal.Chunk{
		Identifier: goriffa.FourCCData,
		Data:       make([]byte, 3),
	}

	// 4 bytes of identifier +
	// 4 bytes of chunk size field +
	// n bytes of data +
	// m padding bytes
	assert.Equal(t, int64(8+len(c.Data)+len(c.Data)%2), c.ByteLength())
}
