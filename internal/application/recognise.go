package application

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"net/url"

	"github.com/speechly/slu-client/pkg/audio"
	"github.com/speechly/slu-client/pkg/audio/wav"
	"github.com/speechly/slu-client/pkg/logger"
	"github.com/speechly/slu-client/pkg/speechly"
	"github.com/speechly/slu-client/pkg/speechly/slu"
)

// RecogniseMicrophone uses Speechly SLU API to recognise audio from the microphone.
func RecogniseMicrophone(
	ctx context.Context, cfg Config, fmt audio.Format, token speechly.AccessToken, dst io.Writer,
	enableTentative bool, bufSize int, log logger.Logger,
) error {
	rec, err := audio.NewRecordStream(fmt, binary.LittleEndian, bufSize, log)
	if err != nil {
		return err
	}

	c := slu.Config{
		NumChannels:     fmt.NumChannels,
		SampleRateHertz: fmt.SampleRateHertz,
		LanguageCode:    cfg.LanguageCode,
	}

	cli, stream, err := newStream(ctx, cfg.SluURL, token, c, log)
	if err != nil {
		return err
	}

	defer func() {
		closeAndLog(stream, "Error closing SLU stream", log)
		closeAndLog(cli, "Error closing SLU client", log)
	}()

	return recogniseSrc(ctx, stream, rec, dst, true, log)
}

// RecogniseFiles uses Speechly API to recognise audio from WAV files stored on disk.
func RecogniseFiles(
	ctx context.Context, cfg Config, token speechly.AccessToken, paths []string, dst io.Writer,
	enableTentative bool, bufSize int, log logger.Logger,
) error {
	if len(paths[0]) < 1 {
		return nil
	}

	// Need to handle first file differently, because we need to access its format to configure recognition stream.
	read, err := wav.NewFileReader(paths[0], bufSize, binary.LittleEndian)
	if err != nil {
		return err
	}
	defer func() {
		if err := read.Close(); err != nil {
			log.Warn("Error closing WAV file reader", err)
		}
	}()

	f := read.Format()
	c := slu.Config{
		NumChannels:     f.NumChannels,
		SampleRateHertz: f.SampleRateHertz,
		LanguageCode:    cfg.LanguageCode,
	}

	cli, stream, err := newStream(ctx, cfg.SluURL, token, c, log)
	if err != nil {
		return err
	}

	defer func() {
		closeAndLog(stream, "Error closing SLU stream", log)
		closeAndLog(cli, "Error closing SLU client", log)
	}()

	if err := recogniseSrc(ctx, stream, read, dst, enableTentative, log); err != nil {
		return err
	}

	for _, p := range paths[1:] {
		r, err := wav.NewFileReader(p, bufSize, binary.LittleEndian)
		if err != nil {
			return err
		}

		if err := recogniseSrc(ctx, stream, r, dst, enableTentative, log); err != nil {
			return err
		}
	}

	return nil
}

func recogniseSrc(
	ctx context.Context, stream slu.RecogniseStream, read slu.AudioSource, dst io.Writer, tent bool, log logger.Logger,
) error {
	defer closeAndLog(read, "Error closing audio source", log)

	out, err := stream.NewAudioContext(ctx, read, 8)
	if err != nil {
		return err
	}
	defer closeAndLog(out, "Error closing SLU context", log)

	var (
		buf = new(bytes.Buffer)
		enc = json.NewEncoder(buf)
	)

	for {
		res, err := out.Read()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		if !(tent || res.IsFinalised) {
			continue
		}

		if err := enc.Encode(res); err != nil {
			return err
		}

		if _, err := io.Copy(dst, buf); err != nil && err != io.EOF {
			return err
		}

		buf.Reset()
	}
}

func newStream(
	ctx context.Context, u url.URL, t speechly.AccessToken, c slu.Config, log logger.Logger,
) (*slu.Client, slu.RecogniseStream, error) {
	cli, err := slu.NewClient(u, t, log)
	if err != nil {
		return nil, nil, err
	}

	if err := cli.Dial(ctx); err != nil {
		return nil, nil, err
	}

	stream, err := cli.StreamingRecognise(ctx, c)
	if err != nil {
		if err := cli.Close(); err != nil {
			log.Warn("Error stopping SLU client", err)
		}

		return nil, nil, err
	}

	return cli, stream, nil
}

func closeAndLog(c io.Closer, msg string, log logger.Logger) {
	if err := c.Close(); err != nil {
		log.Warn(msg, err)
	}
}
