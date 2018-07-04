package dev

import (
	"github.com/spf13/cobra"
)

// moduleCmd represents the fire command
var moduleCmd = &cobra.Command{
	Use:   "module",
	Short: "Execute operations on development modules",
	Run:   Module,
}

// Module base command
func Module(cmd *cobra.Command, args []string) {
	cmd.Help() // nolint: errcheck
}
