package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"net/url"
	"os"

	"github.com/google/uuid"
	"golang.org/x/text/language"

	"github.com/speechly/slu-client/pkg/audio"
	"github.com/speechly/slu-client/pkg/logger"
	"github.com/speechly/slu-client/pkg/speechly/identity"
	"github.com/speechly/slu-client/pkg/speechly/slu"
)

const (
	bufSize = 32 // Bytes
)

func main() {
	ctx := context.Background()
	log := logger.NewStderrLogger()

	// Endpoint URL.
	u, err := url.Parse("grpc+tls://api.speechly.com")
	ensure(err)

	// App language code, can be found in dashboard.
	lang := language.AmericanEnglish

	// App ID, can be found in dashboard.
	appID, err := uuid.Parse("insert-app-id-here")
	ensure(err)

	// Device ID, must be a valid UUID.
	deviceID, err := uuid.NewRandom()
	ensure(err)

	// Audio format to use for microphone.
	f, err := audio.NewFormat(1, 16000, 16)
	ensure(err)

	// Get Speechly API access token from Speechly Identity API.
	token, err := identity.GetAccessToken(ctx, *u, appID, deviceID, log)
	ensure(err)

	// Start new microphone stream.
	rec, err := audio.NewRecordStream(f, binary.LittleEndian, bufSize, log)
	ensure(err)
	defer rec.Close()

	// Initialise SLU client.
	cli, err := slu.NewClient(*u, token, log)
	ensure(err)

	// Dial the client.
	ensure(cli.Dial(ctx))
	defer cli.Close()

	// Start new recognition stream.
	stream, err := cli.StreamingRecognise(ctx, slu.Config{
		NumChannels:     f.NumChannels,
		SampleRateHertz: f.SampleRateHertz,
		LanguageCode:    lang,
	})
	ensure(err)
	defer stream.Close()

	// Start new audio context from the microphone.
	out, err := stream.NewAudioContext(ctx, rec, bufSize)
	ensure(err)
	defer out.Close()

	var (
		buf = new(bytes.Buffer)
		enc = json.NewEncoder(buf)
	)

	// Read responses in a loop, marshal them to JSON and print to STDOUT.
	for {
		res, err := out.Read()
		if err == io.EOF {
			return
		}

		if err != nil {
			panic(err)
		}

		if err := enc.Encode(res); err != nil {
			panic(err)
		}

		if _, err := io.Copy(os.Stdout, buf); err != nil && err != io.EOF {
			panic(err)
		}

		buf.Reset()
	}
}

func ensure(err error) {
	if err != nil {
		panic(err)
	}
}
