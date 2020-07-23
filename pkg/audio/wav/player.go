package wav

import (
	"encoding/binary"

	"speechly/slu-client/pkg/audio"
	"speechly/slu-client/pkg/logger"
)

// NewFilePlayer returns new audio.Player which uses a WAV reader as an audio source.
// WAV reader will use the file specified by path as its data source.
func NewFilePlayer(path string, bufSize int, ord binary.ByteOrder, l logger.Logger) (*audio.Player, error) {
	r, err := NewFileReader(path, bufSize, ord)
	if err != nil {
		return nil, err
	}

	return audio.NewPlayer(r, l)
}

// NewPlayer returns new audio.Player which uses a WAV reader as an audio source.
func NewPlayer(src Source, bufSize int, ord binary.ByteOrder, l logger.Logger) (*audio.Player, error) {
	r, err := NewReader(src, bufSize, ord)
	if err != nil {
		return nil, err
	}

	return audio.NewPlayer(r, l)
}
