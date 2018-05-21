package main

import (
	"fmt"

	"github.com/lawrencegripper/ion/internal/app/sidecar"
	"github.com/spf13/cobra"
)

// NewVersionCommand print the sidecar version
func NewVersionCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "version",
		Short: "Get sidecar's version",
		Run:   func(cmd *cobra.Command, args []string) { fmt.Println(sidecar.Version) },
	}
	return cmd
}
