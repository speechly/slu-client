package wav

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
	"sync"

	gaudio "github.com/go-audio/audio"
	"github.com/go-audio/wav"

	"github.com/speechly/slu-client/pkg/audio"
)

// WAV reader errors.
var (
	ErrInvalidFile = errors.New("invalid WAV file")
	ErrShortWrite  = errors.New("short write")
)

// Source is an interface for audio data source.
type Source interface {
	io.ReadSeeker
	io.Closer
}

// Reader is a WAV file reader that can read serialised WAV files (or other io.Reader sources).
// Reader implements audio.Source and io.WriterTo interfaces.
// Data can be read either by calling Read and accessing the returned audio.Buffer, or using WriteTo.
// In the latter case, data will be serialised to binary using specified byte order.
type Reader struct {
	fmt      audio.Format
	closed   sync.Once
	closeErr error
	intBuf   gaudio.IntBuffer
	buf      audio.Buffer
	enc      binary.ByteOrder
	src      Source
	dec      *wav.Decoder
}

// NewFileReader returns a new Reader that has a file specified by path set as its source.
// It's a convenience wrapper around the following code:
//
// f, err := os.Open(path)
// if err != nil {
// 	return err
// }
//
// reader, err := wav.NewReader(f, bufSize, ord)
func NewFileReader(path string, bufSize int, ord binary.ByteOrder) (*Reader, error) {
	// nolint: gosec // It's expected that the end user of this package ensures the path is safe.
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return NewReader(file, bufSize, ord)
}

// NewReader returns a new Reader that will read WAV data from specified src.
// bufSize controls the size of internal buffer that will be used for reading the data,
// and ord tells the reader which byte order to use for decoding data from src.
func NewReader(src Source, bufSize int, ord binary.ByteOrder) (*Reader, error) {
	dec := wav.NewDecoder(src)
	if !dec.IsValidFile() {
		return nil, ErrInvalidFile
	}

	if !dec.WasPCMAccessed() {
		if err := dec.FwdToPCM(); err != nil {
			return nil, err
		}
	}

	f := dec.Format()
	fmt, err := audio.NewFormat(f.NumChannels, f.SampleRate, int(dec.BitDepth))
	if err != nil {
		return nil, err
	}

	buf, err := audio.NewBuffer(fmt.BitDepth, bufSize)
	if err != nil {
		return nil, err
	}

	return &Reader{
		enc: ord,
		fmt: fmt,
		buf: buf,
		dec: dec,
		src: src,
		intBuf: gaudio.IntBuffer{
			Format: f,
			Data:   make([]int, bufSize),
		},
	}, nil
}

// Format returns the audio format of the data.
func (r *Reader) Format() audio.Format {
	return r.fmt
}

// Buffer returns a copy of buffer used for reading the data.
// Returned buffer can be used for calling ReadBuffer,
// since it's guaranteed to have the same parameters (e.g. bit depth).
func (r *Reader) Buffer() audio.Buffer {
	return r.buf.Clone()
}

// Close closes the reader by closing the underlying src.
// Close can be called multiple times, but only the first time it will actually close src.
// All following calls will simply return the error returned by src.Close, if any.
func (r *Reader) Close() error {
	r.closed.Do(func() {
		r.closeErr = r.src.Close()
	})

	return r.closeErr
}

// ReadBuffer reads and decodes next chunk of data from src, into b.
func (r *Reader) ReadBuffer(b audio.Buffer) (int, error) {
	n, err := r.readNext()
	if err != nil && err != io.EOF {
		return 0, err
	}

	if bn, err := b.Write(r.intBuf.Data, r.fmt.BitDepth); err != nil {
		return 0, err
	} else if bn != n {
		return 0, ErrShortWrite
	}

	return n, err
}

// WriteTo implements io.WriterTo interface.
// It reads data from internal audio buffer, encodes it using Reader byte order and writes it into w.
// It will return io.EOF when the underlying buffer and src have been exhausted.
func (r *Reader) WriteTo(w io.Writer) (int64, error) {
	n, err := r.readNext()
	if err != nil && err != io.EOF {
		return 0, err
	}

	if bn, err := r.buf.Write(r.intBuf.Data, r.fmt.BitDepth); err != nil {
		return 0, err
	} else if bn != n {
		return 0, ErrShortWrite
	}

	if _, err := r.buf.Encode(r.enc, w); err != nil {
		return 0, err
	}

	return int64(n), err
}

func (r *Reader) readNext() (int, error) {
	n, err := r.dec.PCMBuffer(&r.intBuf)

	// For some reason short reads don't return EOF, so rig it here manually.
	if err == nil && (n == 0 || n < len(r.intBuf.Data)) {
		err = io.EOF
	}

	r.intBuf.Data = r.intBuf.Data[:n]

	return n, err
}
