package audio

import (
	"context"

	"github.com/gordonklaus/portaudio"
	"github.com/hashicorp/go-multierror"

	"speechly/slu-client/pkg/logger"
)

// Recorder is an audio recorder that reads data from default OS audio input and writes data to specified  destination.
// It is based on portaudio, so it will use whatever audio stack implementation
// that portaudio implements for current OS.
// It is not safe for concurrent use.
type Recorder struct {
	stream  *portaudio.Stream
	dst     Sink
	buf     Buffer
	log     logger.Logger
	done    chan struct{}
	doneAck chan struct{}
}

// NewRecorder returns a new Recorder that will write audio to dst.
func NewRecorder(dst Sink, log logger.Logger) (*Recorder, error) {
	err := portaudio.Initialize()
	if err != nil {
		return nil, err
	}

	var (
		fmt = dst.Format()
		buf = dst.Buffer()
	)

	stream, err := portaudio.OpenDefaultStream(
		int(fmt.NumChannels), 0, float64(fmt.SampleRateHertz), buf.Size(), buf.Data(),
	)
	if err != nil {
		if err := portaudio.Terminate(); err != nil {
			log.Warn("Error terminating portaudio", err)
		}

		return nil, err
	}

	return &Recorder{
		stream:  stream,
		dst:     dst,
		buf:     buf,
		done:    make(chan struct{}),
		doneAck: make(chan struct{}),
	}, nil
}

// Record records audio and writes it to dst within given context.
func (r *Recorder) Record(ctx context.Context) error {
	if err := r.stream.Start(); err != nil {
		return err
	}

	defer func() {
		if err := r.stream.Stop(); err != nil {
			r.log.Warn("Error stopping audio stream", err)
		}

		close(r.doneAck)
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-r.done:
			return nil
		default:
			if err := r.stream.Read(); err != nil {
				return err
			}

			if _, err := r.dst.WriteBuffer(r.buf); err != nil {
				return err
			}
		}
	}
}

// Close closes the recorder by closing the destination, the audio stream and terminating the audio stack.
// Close MUST be called before program exits,
// otherwise the audio devices of the OS may be unusable until the audio system is restarted.
func (r *Recorder) Close() error {
	// Tell `Record` to exit and wait for ack.
	close(r.done)
	<-r.doneAck

	errs := &multierror.Error{}

	if err := r.stream.Close(); err != nil {
		errs = multierror.Append(errs, err)
	}

	if err := r.dst.Close(); err != nil {
		errs = multierror.Append(errs, err)
	}

	if err := portaudio.Terminate(); err != nil {
		errs = multierror.Append(errs, err)
	}

	return errs.ErrorOrNil()
}
