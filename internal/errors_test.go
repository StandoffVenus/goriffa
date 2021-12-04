package internal_test

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/standoffvenus/goriffa/internal"
	"github.com/stretchr/testify/assert"
)

func TestWrap(t *testing.T) {
	err := errors.New("original")
	wrapped := internal.Wrap("new", err)

	assert.Equal(t, "new", wrapped.Error())
	assert.ErrorIs(t, wrapped, err)
	assert.Equal(t, err, errors.Unwrap(wrapped))
}

func TestPanicOnError(t *testing.T) {
	err := errors.New("error")

	confirmPanics(t, err.Error(), func() { internal.Panic(err) })
}

func TestPanicOnString(t *testing.T) {
	str := "error"

	confirmPanics(t, str, func() { internal.Panic(str) })
}

func TestPanicGeneric(t *testing.T) {
	var i int64 = 42

	confirmPanics(t, strconv.FormatInt(i, 10), func() { internal.Panic(i) })
}

func confirmPanics(t *testing.T, expectingString string, testingFunc func()) {
	defer func() {
		v := recover()
		assert.NotNil(t, v, "Expected function to panic, but received nil recovery value!")

		err, _ := v.(error)
		if assert.Error(t, err) {
			assert.Equal(t, fmt.Sprintf("goriffa: %s", expectingString), err.Error())
		}
	}()

	testingFunc()
}
