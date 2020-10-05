package slu

import (
	"context"
	"io"
	"sync"

	sluv1 "github.com/speechly/api/go/speechly/slu/v1"

	"github.com/speechly/slu-client/pkg/logger"
)

// AudioSource is the interface that represents the audio data source
// that is sent to SLU API within a particular audio context.
type AudioSource interface {
	io.WriterTo
	io.Closer
}

// RecogniseStream is a single SLU recognition stream, which maps to a gRPC stream.
// Since one stream can have multiple contexts RecogniseStream provides an API for launching these.
// However, only a single audio context can be active at a time, which is controlled and guaranteed by the stream.
type RecogniseStream interface {
	// NewAudioContext starts a new audio context by sending a START even to SLU API.
	// If there is already an audio context running,
	// this will block until the running context is stopped, or the stream closed.
	NewAudioContext(context.Context, AudioSource, int) (AudioContextHandler, error)

	// Close closes the stream by closing the sending part of gRPC stream.
	// It will wait for current audio context (if any) to be stopped, before closing the stream.
	Close() error
}

type stream struct {
	stream sluv1.SLU_StreamClient
	log    logger.Logger
	lock   sync.Mutex
}

func newStream(str sluv1.SLU_StreamClient, f Config, log logger.Logger) (*stream, error) {
	if err := str.Send(&sluv1.SLURequest{
		StreamingRequest: &sluv1.SLURequest_Config{
			Config: &sluv1.SLUConfig{
				Encoding:        sluv1.SLUConfig_LINEAR16,
				Channels:        f.NumChannels,
				SampleRateHertz: f.SampleRateHertz,
				LanguageCode:    f.LanguageCode.String(),
			},
		},
	}); err != nil {
		if err := str.CloseSend(); err != nil {
			log.Warn("error closing recognition stream", err)
		}

		return nil, err
	}

	return &stream{
		stream: str,
		log:    log,
	}, nil
}

func (s *stream) NewAudioContext(ctx context.Context, src AudioSource, chanSize int) (AudioContextHandler, error) {
	s.lock.Lock() // Wait for previous context to exit.

	return newCtxHandler(ctx, s.stream, src, chanSize, s.log, func() {
		s.lock.Unlock() // Notify that context is done.
	})
}

func (s *stream) Close() error {
	s.lock.Lock() // Wait for current context to exit (if any).
	defer s.lock.Unlock()
	return s.stream.CloseSend()
}
