package main

import (
	"fmt"
	"os"

	"github.com/speechly/slu-client/cmd/command"
)

func main() {
	if err := command.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		os.Exit(1)
	}
}
