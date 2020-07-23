package audio

import (
	"errors"
	"io"
)

var (
	// ErrInvalidBitDepth is returned when provided BitDepth value is not supported.
	ErrInvalidBitDepth = errors.New("unsupported bit depth")
)

// BitDepth represents the bit depth of encoded audio.
type BitDepth int8

// Possible values for BitDepth are 8, 16, 32 and 64.
const (
	BitDepthUndefined = BitDepth(0)
	BitDepth8         = BitDepth(8)
	BitDepth16        = BitDepth(16)
	BitDepth32        = BitDepth(32)
	BitDepth64        = BitDepth(64)
)

// Format represents the format of audio data.
type Format struct {
	NumChannels     int32
	SampleRateHertz int32
	BitDepth        BitDepth
}

// NewFormat returns new Format with specified number of channels, sample rate and bit depth.
// If provided depth is not valid, an error is returned.
func NewFormat(numChannels, sampleRate, depth int) (Format, error) {
	var d BitDepth
	switch depth {
	case 8:
		d = BitDepth8
	case 16:
		d = BitDepth16
	case 32:
		d = BitDepth32
	case 64:
		d = BitDepth64
	default:
		return Format{}, ErrInvalidBitDepth
	}

	return Format{
		NumChannels:     int32(numChannels),
		SampleRateHertz: int32(sampleRate),
		BitDepth:        d,
	}, nil
}

// Container represents an audio container that holds audio data in a Buffer with a specific Format.
type Container interface {
	Format() Format
	Buffer() Buffer
}

// Source is an audio container that can be read from using ReadBuffer.
type Source interface {
	Container
	io.Closer
	ReadBuffer(Buffer) (int, error)
}

// Sink is an audio container that can be written to using WriteBuffer.
type Sink interface {
	Container
	io.Closer
	WriteBuffer(Buffer) (int, error)
}

// SourceSink is an audio container that can be both written to and read from.
type SourceSink interface {
	Container
	Source
	Sink
}
