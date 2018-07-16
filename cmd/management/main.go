package main

import (
	"os"
)

func main() {
	cmd := NewManagementCommand()

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
