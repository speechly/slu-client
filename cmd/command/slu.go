package command

import (
	"context"
	goos "os"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"

	"speechly/slu-client/internal/application"
	"speechly/slu-client/internal/os"
	"speechly/slu-client/pkg/audio"
	"speechly/slu-client/pkg/speechly"
)

var (
	apiToken        speechly.AccessToken
	enableTentative bool
)

var sluCmd = &cobra.Command{
	Use:   "slu",
	Short: "Interact with Speechly SLU API",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		checkConfig(cmd, args)
		setToken(cmd, args)
	},
}

var streamCmd = &cobra.Command{
	Use:   "stream",
	Short: "Stream audio from microphone to SLU API",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Starting microphone streaming...")

		err := os.WithSignal(cmd.Context(), func(ctx context.Context) error {
			log.Info("Started microphone streaming, press Ctrl+C to finish.")
			audioFmt, err := audio.NewFormat(defaultChanCount, defaultSampleRate, defaultBitDepth)
			if err != nil {
				return err
			}

			return application.RecogniseMicrophone(
				ctx, config, audioFmt, apiToken, goos.Stdout, enableTentative, bufferSize, log,
			)
		}, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

		ensure(err)
		log.Info("Microphone streaming finished!")
	},
}

var uploadCmd = &cobra.Command{
	Use:   "upload file1 file2 ... fileN",
	Short: "Upload a WAV file to SLU API",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Starting file upload...")

		err := os.WithSignal(cmd.Context(), func(ctx context.Context) error {
			paths, err := normalisePaths(args)
			if err != nil {
				return err
			}

			return application.RecogniseFiles(ctx, config, apiToken, paths, goos.Stdout, enableTentative, bufferSize, log)
		}, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

		ensure(err)
		log.Info("File upload finished!")
	},
}

func init() {
	sluCmd.PersistentFlags().BoolVarP(&enableTentative, "enable_tentative", "t", false, "output tentative context states")

	sluCmd.AddCommand(uploadCmd, streamCmd)
	rootCmd.AddCommand(sluCmd)
}

func normalisePaths(paths []string) ([]string, error) {
	p := make([]string, 0, len(paths))

	for _, v := range paths {
		a, err := filepath.Abs(v)
		if err != nil {
			return nil, err
		}

		p = append(p, a)
	}

	return p, nil
}
