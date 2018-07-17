package dev

import (
	// "github.com/lawrencegripper/ion/internal/pkg/servicebus"
	"github.com/spf13/cobra"
)

// eventCmd represents the fire command
var eventCmd = &cobra.Command{
	Use:   "event",
	Short: "Execute operations on development events",
	Run:   Event,
}

//Event root for event cmds
func Event(cmd *cobra.Command, args []string) {
	cmd.Help() // nolint: errcheck
}
