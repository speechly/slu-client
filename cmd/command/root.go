package command

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/speechly/slu-client/internal/application"
	"github.com/speechly/slu-client/pkg/logger"
)

const (
	defaultSampleRate = 16000 // 16 kHz
	defaultBitDepth   = 16    // 16 bit
	defaultChanCount  = 1     // Mono
)

var (
	bufferSize     int
	enableDebug    bool
	configFilePath string
	appID          string
	deviceID       string
	languageCode   string
	sluURL         string
	identityURL    string
)

var log logger.Logger

var rootCmd = &cobra.Command{
	Use:   "speechly-slu",
	Short: "CLI program for working with Speechly SLU API",
}

func Execute() error { // nolint: golint
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(setup)

	rootCmd.PersistentFlags().IntVarP(&bufferSize, "buffer_size", "b", 2048, "Size of memory buffer to use (in bytes).")
	rootCmd.PersistentFlags().StringVarP(&configFilePath, "config", "c", "", "Config file (default $HOME/.speechly).")
	rootCmd.PersistentFlags().BoolVar(&enableDebug, "debug", false, "Enable debug output.")
	rootCmd.PersistentFlags().StringVar(&sluURL, configKeySluURL, "", configDescSluURL)
	rootCmd.PersistentFlags().StringVar(&identityURL, configKeyIdentityURL, "", configDescIdentityURL)
	rootCmd.PersistentFlags().StringVarP(&appID, configKeyAppID, "a", "", configDescAppID)
	rootCmd.PersistentFlags().StringVarP(&deviceID, configKeyDeviceID, "d", "", configDescDeviceID)
	rootCmd.PersistentFlags().StringVarP(&languageCode, configKeyLanguageCode, "l", "", configDescLanguageCode)
}

func setup() {
	level := logrus.InfoLevel
	if enableDebug {
		level = logrus.DebugLevel
	}

	log = application.NewLogger(os.Stderr, level)

	viper.SetDefault(configKeySluURL, configDefaultSluURL)
	viper.SetDefault(configKeyIdentityURL, configDefaultIdentityURL)

	ensure(viper.BindPFlag(configKeySluURL, rootCmd.PersistentFlags().Lookup(configKeySluURL)))
	ensure(viper.BindPFlag(configKeyIdentityURL, rootCmd.PersistentFlags().Lookup(configKeyIdentityURL)))
	ensure(viper.BindPFlag(configKeyAppID, rootCmd.PersistentFlags().Lookup(configKeyAppID)))
	ensure(viper.BindPFlag(configKeyDeviceID, rootCmd.PersistentFlags().Lookup(configKeyDeviceID)))
	ensure(viper.BindPFlag(configKeyLanguageCode, rootCmd.PersistentFlags().Lookup(configKeyLanguageCode)))

	if configFilePath != "" {
		viper.SetConfigFile(configFilePath)
	} else {
		viper.AddConfigPath(getConfigDir())
		viper.SetConfigName(configFileName)
	}

	viper.SetConfigType(configFileFormat)

	if err := viper.ReadInConfig(); err != nil {
		log.Warnf("Error loading config file: '%s'", err)
		return
	}

	if err := parseConfig(); err != nil {
		log.Warnf("Error parsing config: '%s', proceeding without config...", err)
	}
}

func ensure(err error) {
	if err == nil {
		return
	}

	if log != nil {
		log.Warnf("Error: %s", err)
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	}

	os.Exit(1)
}
