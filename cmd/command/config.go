package command

import (
	"context"
	"fmt"
	goos "os"
	"path"
	"sort"
	"strings"
	"syscall"
	"text/tabwriter"

	"github.com/google/uuid"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/speechly/slu-client/internal/application"
	"github.com/speechly/slu-client/internal/os"
)

const (
	configKeySluURL       = "slu_url"
	configKeyIdentityURL  = "identity_url"
	configKeyAppID        = "app_id"
	configKeyDeviceID     = "device_id"
	configKeyLanguageCode = "language_code"

	configDefaultSluURL      = "grpc+tls://api.speechly.com"
	configDefaultIdentityURL = "grpc+tls://api.speechly.com"

	configDescSluURL       = "Speechly SLU API URL. Scheme must be 'grpc+tls://' for TLS URL and 'grpc://' for non-TLS URL."      // nolint: lll
	configDescIdentityURL  = "Speechly Identity API URL. Scheme must be 'grpc+tls://' for TLS URL and 'grpc://' for non-TLS URL." // nolint: lll
	configDescAppID        = "Speechly application identifier, must be a valid UUIDv4."
	configDescDeviceID     = "Device identifier, must be a valid UUIDv4."
	configDescLanguageCode = "Speechly application language code, must be an IETF language tag (e.g. 'en-US')."

	configFileName     = "config"
	configFileFormat   = "json"
	configFileFullName = configFileName + "." + configFileFormat
)

var (
	config          = application.Config{}
	validConfigKeys = map[string]string{
		configKeySluURL:       configDescSluURL,
		configKeyIdentityURL:  configDescIdentityURL,
		configKeyAppID:        configDescAppID,
		configKeyDeviceID:     configDescDeviceID,
		configKeyLanguageCode: configDescLanguageCode,
	}
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration (app ID, language code, etc.)",
}

var configPrintCmd = &cobra.Command{
	Use:    "print",
	Short:  "Print current configuration",
	PreRun: checkConfig,
	Run: func(cmd *cobra.Command, args []string) {
		err := os.WithSignal(cmd.Context(), func(ctx context.Context) error {
			// Sort config keys
			ks := make([]string, 0, len(validConfigKeys))
			for k := range validConfigKeys {
				ks = append(ks, k)
			}
			sort.Strings(ks)

			t := tabwriter.NewWriter(goos.Stdout, 4, 0, 1, ' ', 0)
			for _, k := range ks {
				fmt.Fprintf(t, "%s: \t%s\n", k, viper.GetString(k))
			}
			return t.Flush()
		}, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

		ensure(err)
	},
}

var configGenerateCmd = &cobra.Command{
	Use:     "generate",
	Short:   "Generate configuration file in interactive mode",
	PostRun: removeCachedToken,
	Run: func(cmd *cobra.Command, args []string) {
		err := os.WithSignal(cmd.Context(), func(ctx context.Context) error {
			var res string

			fmt.Print("Speechly app ID: ")
			if _, err := fmt.Scanln(&res); err != nil {
				return err
			}
			viper.Set(configKeyAppID, res)

			fmt.Print("Speechly app language: ")
			if _, err := fmt.Scanln(&res); err != nil {
				return err
			}
			viper.Set(configKeyLanguageCode, res)

			id, err := uuid.NewRandom()
			if err != nil {
				return err
			}

			viper.Set(configKeyDeviceID, id.String())
			viper.Set(configKeySluURL, configDefaultSluURL)
			viper.Set(configKeyIdentityURL, configDefaultIdentityURL)

			if err := parseConfig(); err != nil {
				return err
			}

			if err := touchConfigFile(); err != nil {
				return err
			}

			return viper.WriteConfig()
		}, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

		log.Info("Config generated successfully!")
		ensure(err)
	},
}

var configUpdateCmd = &cobra.Command{
	Use:     "update key value",
	Short:   "Update a specific property in configuration file",
	PreRun:  checkConfig,
	PostRun: removeCachedToken,
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		err := os.WithSignal(cmd.Context(), func(ctx context.Context) error {
			if _, ok := validConfigKeys[args[0]]; !ok {
				return fmt.Errorf(
					"invalid config key '%s', valid keys are:\n%s",
					args[0],
					configKeysHelpString(),
				)
			}

			viper.Set(args[0], args[1])

			if err := parseConfig(); err != nil {
				return err
			}

			if err := viper.WriteConfig(); err != nil {
				return err
			}

			return nil
		}, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

		log.Info("Config updated successfully!")
		ensure(err)
	},
}

func init() {
	configCmd.AddCommand(configGenerateCmd, configUpdateCmd, configPrintCmd)
	rootCmd.AddCommand(configCmd)
}

func checkConfig(cmd *cobra.Command, args []string) {
	if !config.IsValid() {
		if len(goos.Args) > 0 {
			cmd.PrintErrf("Cannot proceed without valid config, please generate one using '%s config generate'!\n", goos.Args[0])
		} else {
			cmd.Println("Cannot proceed without valid config, please generate one using 'config generate'!")
		}

		goos.Exit(1)
	}
}

func parseConfig() error {
	return config.Parse(
		viper.GetString(configKeySluURL),
		viper.GetString(configKeyIdentityURL),
		viper.GetString(configKeyAppID),
		viper.GetString(configKeyDeviceID),
		viper.GetString(configKeyLanguageCode),
	)
}

func touchConfigFile() error {
	dirname := getConfigDir()
	if err := goos.Mkdir(dirname, 0750); err != nil && !goos.IsExist(err) {
		return err
	}

	file, err := goos.Create(path.Join(dirname, configFileFullName))
	if err != nil && !goos.IsExist(err) {
		return err
	}

	return file.Close()
}

func getConfigDir() string {
	home, err := homedir.Dir()
	ensure(err)

	return path.Join(home, ".speechly/")
}

func configKeysHelpString() string {
	b := strings.Builder{}

	for key, desc := range validConfigKeys {
		b.WriteString(fmt.Sprintf("* %s\t%s\n", key, desc))
	}

	return b.String()
}
