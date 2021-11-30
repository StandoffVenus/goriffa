package internal

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const wordLength int64 = 2

// ErrBufferUnderflow can occur when a Read() reads
// fewer bytes than required.
//
// There doesn't seem to be an analog to io.ErrShortWrite
// in the standard library.
var ErrBufferUnderflow error = errors.New("read fewer bytes than expected")

// PaddedLength will return the length of a
// properly padded slice if its original
// length was "n".
func PaddedLength(n int64) int64 {
	if mod := n % wordLength; mod != 0 {
		return n + (wordLength - mod)
	}

	return n
}

// Pad will pad the byte slice as necessary
// to the nearest word with 0's. For instance:
//  []byte{1} -> []byte{1, 0}
func Pad(b []byte) []byte {
	if mod := int64(len(b)) % wordLength; mod != 0 {
		return append(b, make([]byte, (wordLength-mod))...)
	}

	return b
}

// Write will write each byte slice individually, returning
// the total number of bytes written.
// If a write error occurs, all would-be-subsequent writes
// do not occur and the returned integer equals how many bytes
// were written until the error.
// If any write falls short of writing the entire byte slice,
// io.ErrShortWrite is returned.
func Write(w io.Writer, content ...[]byte) (int64, error) {
	var total int64
	for _, slice := range content {
		n, err := w.Write(slice)
		total += int64(n)

		if err != nil {
			return total, err
		}

		if n < len(slice) {
			return total, io.ErrShortWrite
		}
	}

	return total, nil
}

// Read will read into each byte slice individually,
// returning the total number of bytes read.
// If a read error occurs, all would-be-subsequent reads
// do not occur and the returned integer equals how many
// bytes were written until the error.
// If any read falls short of reading into entire byte slice,
// ErrBufferUnderflow is returned.
func Read(r io.Reader, content ...[]byte) (int64, error) {
	var total int64
	for _, slice := range content {
		n, err := r.Read(slice)
		total += int64(n)

		if err != nil {
			return total, err
		}

		if n < len(slice) {
			return total, ErrBufferUnderflow
		}
	}

	return total, nil
}

// Copy will try to copy all bytes from "src"
// into "dst." If the number of bytes copied
// is less than len(src), io.ErrShortBuffer
// is returned.
func Copy(dst, src []byte) error {
	if n := copy(dst, src); n < len(src) {
		return io.ErrShortBuffer
	}

	return nil
}

// WriteAt will write bytes "b" to provided
// writer "w" at offset "offset", returning
// the number of bytes written.
// If w.WriteAt returns an error, it will
// be returned.
// If the total bytes written to w < len(b),
// io.ErrShortWrite is returned.
func WriteAt(w io.WriterAt, b []byte, offset int64) (int, error) {
	n, err := w.WriteAt(b, offset)
	if err != nil {
		return n, err
	}

	if n < len(b) {
		return n, io.ErrShortWrite
	}

	return n, nil
}

// Must4Byte will panic if the provided byte slice is
// not 4 bytes long and can't be copied to a 4-byte
// array.
func Must4Byte(b []byte) [4]byte {
	if len(b) != 4 {
		panic(fmt.Errorf("wrong number of bytes: %d", len(b)))
	}

	var array [4]byte
	if n := copy(array[:], b); n != 4 {
		panic(fmt.Errorf("wrong number of bytes copied: %d", n))
	}

	return array
}

// StringMust4Byte will panic if the provided string is
// not 4 bytes long and can't be copied to a 4-byte
// array.
func StringMust4Byte(s string) [4]byte {
	return Must4Byte([]byte(s))
}

// ReadLittleEndianUInt16 will read a uint16 from
// the reader. If an error is returned, the function
// will panic.
func ReadLittleEndianUInt16(r io.ByteReader) uint16 {
	bytes := make([]byte, 0, 2)
	for i := 0; i < cap(bytes); i++ {
		b, err := r.ReadByte()
		if err != nil {
			panic(err)
		}
		bytes = append(bytes, b)
	}

	return binary.LittleEndian.Uint16(bytes)
}

// ReadLittleEndianUInt32 will read a uint32 from
// the reader. If an error is returned, the function
// will panic.
func ReadLittleEndianUInt32(r io.ByteReader) uint32 {
	bytes := make([]byte, 0, 4)
	for i := 0; i < cap(bytes); i++ {
		b, err := r.ReadByte()
		if err != nil {
			panic(err)
		}
		bytes = append(bytes, b)
	}

	return binary.LittleEndian.Uint32(bytes)
}

// LittleEndianUInt16Bytes is shorthand for
//	var uint16Bytes [2]byte
//	binary.LittleEndian.PutUint16(uint16Bytes[:], u)
func LittleEndianUInt16Bytes(u uint16) []byte {
	var uint16Bytes [2]byte
	binary.LittleEndian.PutUint16(uint16Bytes[:], u)

	return uint16Bytes[:]
}

// LittleEndianUInt32Bytes is shorthand for
//	var uint32Bytes [4]byte
//	binary.LittleEndian.PutUint32(uint32Bytes[:], u)
func LittleEndianUInt32Bytes(u uint32) []byte {
	var uint32Bytes [4]byte
	binary.LittleEndian.PutUint32(uint32Bytes[:], u)

	return uint32Bytes[:]
}
