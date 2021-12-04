package internal

import (
	"errors"
	"fmt"
)

type wrappingError struct {
	message  string
	wrapping error
}

var _ interface{ Unwrap() error } = wrappingError{}

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

// Wrap returns a new error with the given message,
// such that when Wrap() is called, the provided
// error is returned; in other words, Wrap will
// return an error that wraps "err" with a new
// message ("msg")
//
// Passing a nil error is not allowed and will
// result in panic.
func Wrap(msg string, err error) error {
	if err != nil {
		return wrappingError{
			message:  msg,
			wrapping: err,
		}
	}

	Panic("cannot wrap nil error")

	return nil // Satisifies compiler.
}

// Panic will cause the application to panic
// with the provided value, ensuring it will
// some how be prefix with "goriffa".
//
// If the value is an error, then it is wrapped
// via Wrap and the panic value will be the
// wrapped error.
//
// If the value is a string, then the panic
// value is said string with "goriffa: "
// prepended.
//
// Otherwise, the panic value is equivalent
// to:
//  fmt.Sprintf("goriffa: %v", v)
func Panic(v interface{}) {
	const prefix = "goriffa: "
	switch concrete := v.(type) {
	case error:
		panic(Wrap(prefix+concrete.Error(), concrete))
	case string:
		panic(fmt.Errorf("%s%s", prefix, concrete))
	default:
		panic(fmt.Errorf("%s%v", prefix, concrete))
	}
}

func (e wrappingError) Unwrap() error {
	return e.wrapping
}

func (e wrappingError) Error() string {
	return e.message
}
