package dev

import (
	// "github.com/lawrencegripper/ion/internal/pkg/servicebus"
	"github.com/spf13/cobra"
)

var eventName string
var serviceBusConnectionString string

// moduleCmd represents the fire command
var moduleCmd = &cobra.Command{
	Use:   "module",
	Short: "Execute operations on development modules",
	Run:   Module,
}

// Module
func Module(cmd *cobra.Command, args []string) {
	cmd.Help() // nolint: errcheck
}
