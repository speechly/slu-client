package wav

import (
	"os"

	"github.com/speechly/slu-client/pkg/audio"
	"github.com/speechly/slu-client/pkg/logger"
)

// NewFileRecorder returns a new audio.Recorder with a WAV writer set as destination.
// The writer will write to a file specified by path.
func NewFileRecorder(path string, fmt audio.Format, bufSize int, l logger.Logger) (*audio.Recorder, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	return NewRecorder(f, fmt, bufSize, l)
}

// NewRecorder returns a new audio.Recorder with a WAV writer set as destination.
func NewRecorder(dst Sink, fmt audio.Format, bufSize int, l logger.Logger) (*audio.Recorder, error) {
	w, err := NewWriter(dst, fmt, bufSize)
	if err != nil {
		return nil, err
	}

	return audio.NewRecorder(w, l)
}
