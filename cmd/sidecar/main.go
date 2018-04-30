package main

import (
	"os"
)

func main() {
	cmd := NewSidecarCommand()

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
