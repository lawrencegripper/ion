package dev

import (
	"github.com/lawrencegripper/ion/cmd/ionctl/root"

	"github.com/spf13/cobra"
)

// devCmd represents the events command
var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "execute development operations",
}

// Register adds to root command
func Register() {

	// Add module sub commands
	moduleCmd.AddCommand(initCmd)
	moduleCmd.AddCommand(commitCmd)
	moduleCmd.AddCommand(prepareCmd)

	// Add event sub commands
	devCmd.AddCommand(moduleCmd)

	// Add event command to root
	root.RootCmd.AddCommand(devCmd)
}
