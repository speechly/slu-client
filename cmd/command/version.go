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
		fmt.Printf("Build number:\t\t%s\n", application.BuildVersion)
		fmt.Printf("Build author:\t\t%s\n", application.BuildAuthor)
		fmt.Printf("Build timestamp:\t%s\n", application.BuildTime)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
