package audio

import (
	"context"
	"io"

	"github.com/gordonklaus/portaudio"
	"github.com/hashicorp/go-multierror"

	"github.com/speechly/slu-client/pkg/logger"
)

// Player is an audio player that reads data from specified audio source,
// and plays it using OS audio stack through default output device.
// It is based on portaudio, so it will use whatever audio stack implementation
// that portaudio implements for current OS.
// It is not safe for concurrent use.
type Player struct {
	stream  *portaudio.Stream
	src     Source
	buf     Buffer
	log     logger.Logger
	done    chan struct{}
	doneAck chan struct{}
}

// NewPlayer returns a new Player that will use src as source of audio data.
func NewPlayer(src Source, log logger.Logger) (*Player, error) {
	err := portaudio.Initialize()
	if err != nil {
		return nil, err
	}

	var (
		format = src.Format()
		buf    = src.Buffer()
	)

	stream, err := portaudio.OpenDefaultStream(
		0, int(format.NumChannels), float64(format.SampleRateHertz), buf.Size(), buf.Data(),
	)
	if err != nil {
		if err := portaudio.Terminate(); err != nil {
			log.Warn("Error terminating portaudio:", err)
		}

		return nil, err
	}

	return &Player{
		stream:  stream,
		src:     src,
		buf:     buf,
		log:     log,
		done:    make(chan struct{}),
		doneAck: make(chan struct{}),
	}, nil
}

// Play plays audio from src within given context.
func (p *Player) Play(ctx context.Context) error {
	if err := p.stream.Start(); err != nil {
		return err
	}

	defer func() {
		if err := p.stream.Stop(); err != nil {
			p.log.Warn("Error stopping audio stream", err)
		}

		close(p.doneAck)
	}()

	for done := false; !done; {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-p.done:
			return io.EOF
		default:
			_, err := p.src.ReadBuffer(p.buf)
			if err != nil && err != io.EOF {
				return err
			}

			if err == io.EOF {
				done = true
			}

			if err := p.stream.Write(); err != nil {
				return err
			}
		}
	}

	return nil
}

// Close closes the player by closing the source, the audio stream and terminating the audio stack.
// Close MUST be called before program exits,
// otherwise the audio devices of the OS may be unusable until the audio system is restarted.
func (p *Player) Close() error {
	// Tell `Play` to exit and wait for ack.
	close(p.done)
	<-p.doneAck

	errs := &multierror.Error{}

	if err := p.src.Close(); err != nil {
		errs = multierror.Append(errs, err)
	}

	if err := p.stream.Close(); err != nil {
		errs = multierror.Append(errs, err)
	}

	if err := portaudio.Terminate(); err != nil {
		errs = multierror.Append(errs, err)
	}

	return errs.ErrorOrNil()
}
