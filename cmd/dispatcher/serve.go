package main

import (
	"github.com/spf13/cobra"
)

func NewCmdServe() *cobra.Command {
	cmd := &cobra.Command{
		Use: "serve",
		//Short:
		Run: func(cmd *cobra.Command, args []string) {

		},
	}

	return cmd
}
