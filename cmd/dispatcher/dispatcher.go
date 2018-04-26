package main

import (
	"github.com/spf13/cobra"
)

// NewDispatcherCommand return cobra.Command to run ion-disptacher command
func NewDispatcherCommand() *cobra.Command {
	dispatcherCmd := &cobra.Command{
		Use:   "ion-dispatcher",
		Short: "ion-dispatcher: ...",
	}

	dispatcherCmd.AddCommand(NewCmdServe())

	return dispatcherCmd
}
