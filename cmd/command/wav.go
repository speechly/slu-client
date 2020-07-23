package command

import (
	"context"
	"encoding/binary"
	"syscall"

	"github.com/spf13/cobra"

	"speechly/slu-client/internal/os"
	"speechly/slu-client/pkg/audio"
	"speechly/slu-client/pkg/audio/wav"
)

var wavCmd = &cobra.Command{
	Use:   "wav",
	Short: "Interact with WAV audio files",
}

var wavPlayCmd = &cobra.Command{
	Use:   "play filepath",
	Short: "Play a WAV file from disk",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Starting playback...")

		err := os.WithSignal(cmd.Context(), func(ctx context.Context) error {
			p, err := wav.NewFilePlayer(args[0], bufferSize, binary.LittleEndian, log)
			ensure(err)
			defer func() {
				if err := p.Close(); err != nil {
					log.Warn("Error closing WAV player:", err)
				}
			}()

			return p.Play(ctx)
		}, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

		ensure(err)
		log.Info("Playback finished!")
	},
}

var wavRecordCmd = &cobra.Command{
	Use:   "record filepath",
	Short: "Record audio from microphone to a WAV file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Starting recording, press Ctrl+C to finish...")

		err := os.WithSignal(cmd.Context(), func(ctx context.Context) error {
			format, err := audio.NewFormat(defaultChanCount, defaultSampleRate, defaultBitDepth)
			if err != nil {
				return err
			}

			rec, err := wav.NewFileRecorder(args[0], format, bufferSize, log)
			if err != nil {
				return err
			}
			defer func() {
				if err := rec.Close(); err != nil {
					log.Warn("Error closing WAV recorder:", err)
				}
			}()

			return rec.Record(ctx)
		}, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

		if err == context.Canceled {
			log.Info("Recording finished!")
			return
		}

		ensure(err)
	},
}

func init() {
	wavCmd.AddCommand(wavPlayCmd, wavRecordCmd)
	rootCmd.AddCommand(wavCmd)
}
