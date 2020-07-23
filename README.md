# Speechly SLU API client

This repository contains the source code for Speechly SLU API client, written in Go. It also includes functionality to record and play WAV audio files. Recorded files can then be streamed to SLU API, instead of streaming the microphone.

The client can be used as a standalone CLI app or included in other Go projects as a library.

## Installation

The client uses [portaudio](http://www.portaudio.com) for audio I/O, so it needs to be installed. Pre-built binaries for macOS and Linux amd64 arch are automatically compiled on every release - https://github.com/speechly/slu-client/releases, you can download them using e.g. `curl`.

### macOS

Install portaudio with Homebrew and download the binary:

```sh
brew install portaudio
curl -L https://github.com/speechly/slu-client/releases/latest/download/speechly-slu-macos-amd64.tar.gz | tar xz
```

or, if you want a specific version:

```sh
export VERSION="v0.1.0"
brew install portaudio
curl -L https://github.com/speechly/slu-client/releases/latest/download/${VERSION}/speechly-slu-macos-amd64.tar.gz | tar xz
```

### Ubuntu / Debian

Make sure to install `libportaudio2`. Currently only `amd64` arch binaries are pre-built:

```sh
sudo apt-get update && sudo apt-get install libportaudio2
curl -L https://github.com/speechly/slu-client/releases/latest/download/speechly-slu-linux-amd64.tar.gz | tar xz
```

### Building from scratch

Alternatively, you can build it yourself, make sure you have `make` and `go` installed. Minimum support Go version is `1.14`.

#### macOS:

```sh
# Pre-requisites, assuming you have Homebrew installed.
brew install git make go portaudio pkg-config

# Clone the repo
git clone git@github.com:speechly/slu-client.git
cd slu-client

# Build
make build

# Run
./bin/speechly-slu help
```

#### Ubuntu / Debian

```sh
sudo apt-get update && sudo apt-get install git golang make portaudio19-dev libportaudio2

# Clone the repo
git clone git@github.com:speechly/slu-client.git
cd slu-client

# Build
make build

# Run
./bin/speechly-slu help
```

## Usage

### CLI

```sh
# Make sure to generate config file first.
# It will prompt you to enter your app ID and app language, which can be obtained from the dashboard.
speechly-slu config generate

# Print back the config:
speechly-slu config print

# Stream your microphone input to SLU API.
speechly-slu slu stream

# You can record a WAV file and upload it to API if you need to test the same audio with different SLU configurations:
speechly-slu wav record my_wav_file.wav

# Play back the file to check what it sounds like:
speechly-slu wav play my_wav_file.wav

# Upload the file to SLU API:
speechly-slu slu upload my_wav_file.wav
```

More CLI options available, you can override some settings on the fly or fine-tune memory usage, just use the `help $command`:

```
$ speechly-slu help slu
Interact with Speechly SLU API

Usage:
  speechly-slu slu [command]

Available Commands:
  stream      Stream audio from microphone to SLU API
  upload      Upload a WAV file to SLU API

Flags:
  -t, --enable_tentative   output tentative context states
  -h, --help               help for slu

Global Flags:
  -a, --app_id string          Speechly application identifier, must be a valid UUIDv4.
  -b, --buffer_size int        Size of memory buffer to use (in bytes). (default 2048)
  -c, --config string          Config file (default $HOME/.speechly).
      --debug                  Enable debug output.
  -d, --device_id string       Device identifier, must be a valid UUIDv4.
      --identity_url string    Speechly Identity API URL. Scheme must be 'grpc+tls://' for TLS URL and 'grpc://' for non-TLS URL.
  -l, --language_code string   Speechly application language code, must be an IETF language tag (e.g. 'en-US').
      --slu_url string         Speechly SLU API URL. Scheme must be 'grpc+tls://' for TLS URL and 'grpc://' for non-TLS URL.

Use "speechly-slu slu [command] --help" for more information about a command.
```

### Go library

Get it using `go get`:

```sh
$ go get github.com/speechly/slu-client
```

The library provides two main packages - `speechly` and `audio`, which can be used to interact with Speechly APIs or work with audio files.

#### Example

Here's an end-to-end example of how to stream your microphone to Speechly SLU API, you can also check it in [microphone.go](examples/microphone.go):

```golang
package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/google/uuid"
	"golang.org/x/text/language"

	"github.com/speechly/slu-client/pkg/audio"
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
```
