package wav

import (
	"io"

	gaudio "github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/hashicorp/go-multierror"

	"speechly/slu-client/pkg/audio"
)

// Supported WAV encodings.
const (
	EncodingPCM = 1
)

// Sink is an interface for a WAV data source.
type Sink interface {
	io.WriteSeeker
	io.Closer
}

// Writer is a WAV file writer that performs audio data encoding,
// serialisation into binary format and headers.
// Writer implements audio.Sink and io.Closer interfaces.
type Writer struct {
	dst    io.Closer
	fmt    audio.Format
	buf    audio.Buffer
	intBuf gaudio.IntBuffer
	enc    *wav.Encoder
}

// NewWriter returns a new Writer that will write data to specified dst.
// fmt tells the writer what format of audio to use (sample rate, number of channels, etc.).
// and bufSize controls the size of internal buffer that will be used for writing the data.
func NewWriter(dst Sink, fmt audio.Format, bufSize int) (*Writer, error) {
	buf, err := audio.NewBuffer(fmt.BitDepth, bufSize)
	if err != nil {
		return nil, err
	}

	return &Writer{
		fmt: fmt,
		buf: buf,
		dst: dst,
		enc: wav.NewEncoder(dst, int(fmt.SampleRateHertz), int(fmt.BitDepth), int(fmt.NumChannels), EncodingPCM),
		intBuf: gaudio.IntBuffer{
			Data:           make([]int, bufSize),
			SourceBitDepth: int(fmt.BitDepth),
			Format: &gaudio.Format{
				NumChannels: int(fmt.NumChannels),
				SampleRate:  int(fmt.SampleRateHertz),
			},
		},
	}, nil
}

// Format returns the audio format of the writer.
func (w *Writer) Format() audio.Format {
	return w.fmt
}

// Buffer returns a copy of buffer used for writing the data.
// Returned buffer can be used for calling WriteBuffer,
// since it's guaranteed to have the same parameters (e.g. bit depth).
func (w *Writer) Buffer() audio.Buffer {
	return w.buf.Clone()
}

// WriteBuffer writes data from b into the writer.
func (w *Writer) WriteBuffer(b audio.Buffer) (int, error) {
	// Resize to max capacity before reading
	w.intBuf.Data = w.intBuf.Data[:cap(w.intBuf.Data)]

	n, err := b.Read(w.intBuf.Data, w.fmt.BitDepth)
	if err != nil {
		return 0, err
	}

	// Resize to valid length, since encoder.Write doesn't accept number of values to write
	w.intBuf.Data = w.intBuf.Data[:n]

	// Write the buffer to encoder.
	if err := w.enc.Write(&w.intBuf); err != nil {
		return 0, err
	}

	return n, nil
}

// Close closes the writer by closing the encoder and the underlying dst.
func (w *Writer) Close() error {
	errs := &multierror.Error{}

	if err := w.enc.Close(); err != nil {
		errs = multierror.Append(errs, err)
	}

	if err := w.dst.Close(); err != nil {
		errs = multierror.Append(errs, err)
	}

	return errs.ErrorOrNil()
}
