package audio

import (
	"encoding/binary"
	"errors"
	"io"
)

// ErrInvalidBuffer is returned when specified audio buffer is not valid for the operation.
var ErrInvalidBuffer = errors.New("invalid buffer")

// Buffer represents a chunk of audio data of a specific size and bit depth.
// TODO: need to make sure that Write / Read are balanced and do not overwrite.
type Buffer interface {
	// Size returns the size of the underlying slice.
	Size() int

	// BitDepth returns bit depth of the buffer.
	BitDepth() BitDepth

	// Data returns the pointer to the underlying slice.
	Data() interface{}

	// Clone clones the buffer, copying underlying slice.
	Clone() Buffer

	// Write writes provided int slice into the buffer.
	Write([]int, BitDepth) (int, error)

	// Read reads the contents of the buffer into provided slice.
	Read([]int, BitDepth) (int, error)

	// WriteTo writes the contents of the buffer into buf.
	// buf has to have matching bit depth.
	WriteTo(buf Buffer) (int, error)

	// ReadFrom reads the contents of buf into the buffer.
	// buf has to have matching bit depth.
	ReadFrom(buf Buffer) (int, error)

	// Encode writes the contents of the buffer into provided writer.
	// The buffer is responsible for performing resampling.
	Encode(binary.ByteOrder, io.Writer) (int, error)

	// Decode reads the contents of provided reader into the buffer.
	// The buffer is responsible for performing resampling.
	Decode(binary.ByteOrder, io.Reader) (int, error)
}

// NewBuffer returns a new Buffer with specified bit depth and size.
func NewBuffer(d BitDepth, size int) (Buffer, error) {
	switch d {
	case BitDepth8:
		return newInt8Buffer(size), nil
	case BitDepth16:
		return newInt16Buffer(size), nil
	case BitDepth32:
		return newInt32Buffer(size), nil
	case BitDepth64:
		return newInt64Buffer(size), nil
	default:
		return nil, ErrInvalidBitDepth
	}
}

type int8Buffer struct {
	data []int8
	size int
}

func newInt8Buffer(size int) Buffer {
	return int8Buffer{
		data: make([]int8, size),
		size: size,
	}
}

func (b int8Buffer) Size() int {
	return b.size
}

func (b int8Buffer) BitDepth() BitDepth {
	return BitDepth16
}

func (b int8Buffer) Data() interface{} {
	return &b.data
}

func (b int8Buffer) Clone() Buffer {
	data := make([]int8, b.size)
	copy(data, b.data)

	return int8Buffer{
		data: data,
		size: b.size,
	}
}

func (b int8Buffer) Write(buf []int, d BitDepth) (int, error) {
	var (
		dlen = len(b.data)
		blen = len(buf)
	)

	if dlen < blen {
		b.data = append(b.data, make([]int8, blen-dlen)...)
	} else if dlen > blen {
		b.data = b.data[:blen]
	}

	for i := 0; i < len(buf); i++ {
		b.data[i] = int8(buf[i])
	}

	return len(b.data), nil
}

func (b int8Buffer) Read(buf []int, d BitDepth) (int, error) {
	ln := len(buf)
	if len(b.data) < ln {
		ln = len(b.data)
	}

	for i := 0; i < ln; i++ {
		buf[i] = int(b.data[i])
	}

	return ln, nil
}

func (b int8Buffer) WriteTo(buf Buffer) (int, error) {
	// Changing bit depths is not yet supported.
	if b.BitDepth() != buf.BitDepth() {
		return 0, ErrInvalidBitDepth
	}

	s, ok := buf.Data().(*[]int8)
	if !ok {
		return 0, ErrInvalidBuffer
	}

	return copy(*s, b.data), nil
}

func (b int8Buffer) ReadFrom(buf Buffer) (int, error) {
	// Changing bit depths is not yet supported.
	if b.BitDepth() != buf.BitDepth() {
		return 0, ErrInvalidBitDepth
	}

	s, ok := buf.Data().(*[]int8)
	if !ok {
		return 0, ErrInvalidBuffer
	}

	return copy(b.data, *s), nil
}

func (b int8Buffer) Encode(enc binary.ByteOrder, w io.Writer) (int, error) {
	n := 0

	for _, v := range b.data {
		if err := binary.Write(w, enc, v); err != nil {
			return n, err
		}

		n++
	}

	return n, nil
}

func (b int8Buffer) Decode(enc binary.ByteOrder, r io.Reader) (int, error) {
	// Make sure we're at max capacity
	b.data = b.data[:cap(b.data)]

	n := 0

	for i := 0; i < len(b.data); i++ {
		if err := binary.Read(r, enc, &b.data[i]); err != nil {
			return n, err
		}

		n++
	}

	return n, nil
}

type int16Buffer struct {
	data []int16
	size int
}

func newInt16Buffer(size int) Buffer {
	return int16Buffer{
		data: make([]int16, size),
		size: size,
	}
}

func (b int16Buffer) Size() int {
	return b.size
}

func (b int16Buffer) BitDepth() BitDepth {
	return BitDepth16
}

func (b int16Buffer) Data() interface{} {
	return &b.data
}

func (b int16Buffer) Clone() Buffer {
	data := make([]int16, b.size)
	copy(data, b.data)

	return int16Buffer{
		data: data,
		size: b.size,
	}
}

func (b int16Buffer) Write(buf []int, d BitDepth) (int, error) {
	var (
		dlen = len(b.data)
		blen = len(buf)
	)

	if dlen < blen {
		b.data = append(b.data, make([]int16, blen-dlen)...)
	} else if dlen > blen {
		b.data = b.data[:blen]
	}

	for i := 0; i < len(buf); i++ {
		b.data[i] = int16(buf[i])
	}

	return len(b.data), nil
}

func (b int16Buffer) Read(buf []int, d BitDepth) (int, error) {
	ln := len(buf)
	if len(b.data) < ln {
		ln = len(b.data)
	}

	for i := 0; i < ln; i++ {
		buf[i] = int(b.data[i])
	}

	return ln, nil
}

func (b int16Buffer) WriteTo(buf Buffer) (int, error) {
	// Changing bit depths is not yet supported.
	if b.BitDepth() != buf.BitDepth() {
		return 0, ErrInvalidBitDepth
	}

	s, ok := buf.Data().(*[]int16)
	if !ok {
		return 0, ErrInvalidBuffer
	}

	return copy(*s, b.data), nil
}

func (b int16Buffer) ReadFrom(buf Buffer) (int, error) {
	// Changing bit depths is not yet supported.
	if b.BitDepth() != buf.BitDepth() {
		return 0, ErrInvalidBitDepth
	}

	s, ok := buf.Data().(*[]int16)
	if !ok {
		return 0, ErrInvalidBuffer
	}

	return copy(b.data, *s), nil
}

func (b int16Buffer) Encode(enc binary.ByteOrder, w io.Writer) (int, error) {
	n := 0

	for _, v := range b.data {
		if err := binary.Write(w, enc, v); err != nil {
			return n, err
		}

		n++
	}

	return n, nil
}

func (b int16Buffer) Decode(enc binary.ByteOrder, r io.Reader) (int, error) {
	// Make sure we're at max capacity
	b.data = b.data[:cap(b.data)]

	n := 0

	for i := 0; i < len(b.data); i++ {
		if err := binary.Read(r, enc, &b.data[i]); err != nil {
			return n, err
		}

		n++
	}

	return n, nil
}

type int32Buffer struct {
	data []int32
	size int
}

func newInt32Buffer(size int) Buffer {
	return int32Buffer{
		data: make([]int32, size),
		size: size,
	}
}

func (b int32Buffer) Size() int {
	return b.size
}

func (b int32Buffer) BitDepth() BitDepth {
	return BitDepth16
}

func (b int32Buffer) Data() interface{} {
	return &b.data
}

func (b int32Buffer) Clone() Buffer {
	data := make([]int32, b.size)
	copy(data, b.data)

	return int32Buffer{
		data: data,
		size: b.size,
	}
}

func (b int32Buffer) Write(buf []int, d BitDepth) (int, error) {
	var (
		dlen = len(b.data)
		blen = len(buf)
	)

	if dlen < blen {
		b.data = append(b.data, make([]int32, blen-dlen)...)
	} else if dlen > blen {
		b.data = b.data[:blen]
	}

	for i := 0; i < len(buf); i++ {
		b.data[i] = int32(buf[i])
	}

	return len(b.data), nil
}

func (b int32Buffer) Read(buf []int, d BitDepth) (int, error) {
	ln := len(buf)
	if len(b.data) < ln {
		ln = len(b.data)
	}

	for i := 0; i < ln; i++ {
		buf[i] = int(b.data[i])
	}

	return ln, nil
}

func (b int32Buffer) WriteTo(buf Buffer) (int, error) {
	// Changing bit depths is not yet supported.
	if b.BitDepth() != buf.BitDepth() {
		return 0, ErrInvalidBitDepth
	}

	s, ok := buf.Data().(*[]int32)
	if !ok {
		return 0, ErrInvalidBuffer
	}

	return copy(*s, b.data), nil
}

func (b int32Buffer) ReadFrom(buf Buffer) (int, error) {
	// Changing bit depths is not yet supported.
	if b.BitDepth() != buf.BitDepth() {
		return 0, ErrInvalidBitDepth
	}

	s, ok := buf.Data().(*[]int32)
	if !ok {
		return 0, ErrInvalidBuffer
	}

	return copy(b.data, *s), nil
}

func (b int32Buffer) Encode(enc binary.ByteOrder, w io.Writer) (int, error) {
	n := 0

	for _, v := range b.data {
		if err := binary.Write(w, enc, v); err != nil {
			return n, err
		}

		n++
	}

	return n, nil
}

func (b int32Buffer) Decode(enc binary.ByteOrder, r io.Reader) (int, error) {
	// Make sure we're at max capacity
	b.data = b.data[:cap(b.data)]

	n := 0

	for i := 0; i < len(b.data); i++ {
		if err := binary.Read(r, enc, &b.data[i]); err != nil {
			return n, err
		}

		n++
	}

	return n, nil
}

type int64Buffer struct {
	data []int64
	size int
}

func newInt64Buffer(size int) Buffer {
	return int64Buffer{
		data: make([]int64, size),
		size: size,
	}
}

func (b int64Buffer) Size() int {
	return b.size
}

func (b int64Buffer) BitDepth() BitDepth {
	return BitDepth16
}

func (b int64Buffer) Data() interface{} {
	return &b.data
}

func (b int64Buffer) Clone() Buffer {
	data := make([]int64, b.size)
	copy(data, b.data)

	return int64Buffer{
		data: data,
		size: b.size,
	}
}

func (b int64Buffer) Write(buf []int, d BitDepth) (int, error) {
	var (
		dlen = len(b.data)
		blen = len(buf)
	)

	if dlen < blen {
		b.data = append(b.data, make([]int64, blen-dlen)...)
	} else if dlen > blen {
		b.data = b.data[:blen]
	}

	for i := 0; i < len(buf); i++ {
		b.data[i] = int64(buf[i])
	}

	return len(b.data), nil
}

func (b int64Buffer) Read(buf []int, d BitDepth) (int, error) {
	ln := len(buf)
	if len(b.data) < ln {
		ln = len(b.data)
	}

	for i := 0; i < ln; i++ {
		buf[i] = int(b.data[i])
	}

	return ln, nil
}

func (b int64Buffer) WriteTo(buf Buffer) (int, error) {
	// Changing bit depths is not yet supported.
	if b.BitDepth() != buf.BitDepth() {
		return 0, ErrInvalidBitDepth
	}

	s, ok := buf.Data().(*[]int64)
	if !ok {
		return 0, ErrInvalidBuffer
	}

	return copy(*s, b.data), nil
}

func (b int64Buffer) ReadFrom(buf Buffer) (int, error) {
	// Changing bit depths is not yet supported.
	if b.BitDepth() != buf.BitDepth() {
		return 0, ErrInvalidBitDepth
	}

	s, ok := buf.Data().(*[]int64)
	if !ok {
		return 0, ErrInvalidBuffer
	}

	return copy(b.data, *s), nil
}

func (b int64Buffer) Encode(enc binary.ByteOrder, w io.Writer) (int, error) {
	n := 0

	for _, v := range b.data {
		if err := binary.Write(w, enc, v); err != nil {
			return n, err
		}

		n++
	}

	return n, nil
}

func (b int64Buffer) Decode(enc binary.ByteOrder, r io.Reader) (int, error) {
	// Make sure we're at max capacity
	b.data = b.data[:cap(b.data)]

	n := 0

	for i := 0; i < len(b.data); i++ {
		if err := binary.Read(r, enc, &b.data[i]); err != nil {
			return n, err
		}

		n++
	}

	return n, nil
}
