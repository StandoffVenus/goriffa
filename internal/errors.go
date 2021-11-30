package internal

import "errors"

var (
	// ErrClosed represents the case where a reader or
	// writer has been closed.
	ErrClosed error = errors.New("closed")

	// ErrCorrupted represents the case where RIFF data is
	// corrupted.
	ErrCorrupted error = errors.New("corrupted")

	// ErrBadChunk is returned when an attempt is made to
	// parse a chunk but the data is in an invalid format.
	ErrBadChunk error = errors.New("invalid chunk")
)
