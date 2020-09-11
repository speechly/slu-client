package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/speechly/slu-client/internal/application"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print application's version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version:   %s\n", application.BuildVersion)
		fmt.Printf("Timestamp: %s\n", application.BuildTime)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
