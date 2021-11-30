package goriffa

import (
	"github.com/standoffvenus/goriffa/internal"
)

var (
	// ErrClosed represents the case where a reader or
	// writer has been closed.
	ErrClosed error = internal.ErrClosed

	// ErrCorrupted represents the case where RIFF data is
	// corrupted.
	ErrCorrupted error = internal.ErrCorrupted

	// ErrBadChunk is returned when an attempt is made to
	// parse a chunk but the data is in an invalid format.
	ErrBadChunk error = internal.ErrBadChunk
)
