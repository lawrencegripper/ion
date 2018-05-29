package main

import (
	"os"
)

func main() {
	cmd := NewHandlerCommand()

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
