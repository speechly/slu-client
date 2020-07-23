package audio

import (
	"encoding/binary"
	"io"
	"sync"

	"github.com/gordonklaus/portaudio"
	"github.com/hashicorp/go-multierror"

	"github.com/speechly/slu-client/pkg/logger"
)

// RecordStream is an audio stream that implements io.WriterTo interface,
// by using an audio buffer and encoding it into binary data using specified audio format and byte order.
type RecordStream struct {
	stream   *portaudio.Stream
	buf      Buffer
	ord      binary.ByteOrder
	log      logger.Logger
	closed   sync.Once
	closeErr error
}

// NewRecordStream returns a new record stream with specified format, byte order and size of underlying buffer.
func NewRecordStream(fmt Format, ord binary.ByteOrder, bufSize int, log logger.Logger) (*RecordStream, error) {
	buf, err := NewBuffer(fmt.BitDepth, bufSize)
	if err != nil {
		return nil, err
	}

	if err := portaudio.Initialize(); err != nil {
		return nil, err
	}

	stream, err := portaudio.OpenDefaultStream(
		int(fmt.NumChannels), 0, float64(fmt.SampleRateHertz), buf.Size(), buf.Data(),
	)
	if err != nil {
		if err := portaudio.Terminate(); err != nil {
			log.Warn("Error terminating portaudio", log)
		}

		return nil, err
	}

	if err := stream.Start(); err != nil {
		if err := stream.Close(); err != nil {
			log.Warn("Error closing audio stream", log)
		}

		if err := portaudio.Terminate(); err != nil {
			log.Warn("Error terminating portaudio", log)
		}

		return nil, err
	}

	return &RecordStream{
		stream: stream,
		buf:    buf,
		ord:    ord,
		log:    log,
	}, nil
}

// WriteTo implements io.WriterTo by reading from stream and then encoding and writing audio to w.
func (r *RecordStream) WriteTo(w io.Writer) (int64, error) {
	if err := r.stream.Read(); err != nil {
		return 0, err
	}

	n, err := r.buf.Encode(r.ord, w)
	return int64(n), err
}

// Close closes RecordStream by closing the audio stream and terminating the audio stack.
// Close MUST be called before program exits,
// otherwise the audio devices of the OS may be unusable until the audio system is restarted.
func (r *RecordStream) Close() error {
	r.closed.Do(func() {
		errs := &multierror.Error{}

		if err := r.stream.Stop(); err != nil {
			errs = multierror.Append(errs, err)
		}

		if err := r.stream.Close(); err != nil {
			errs = multierror.Append(errs, err)
		}

		if err := portaudio.Terminate(); err != nil {
			errs = multierror.Append(errs, err)
		}

		r.closeErr = errs.ErrorOrNil()
	})

	return r.closeErr
}
