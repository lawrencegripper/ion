package main

import (
	"github.com/lawrencegripper/ion/internal/app/dispatcher"

	"github.com/spf13/cobra"
)

func NewCmdStart() *cobra.Command {
	cmd := &cobra.Command{
		Use: "start",
		Short: "Instanciate the dispatcher to process events"
		Run: func(cmd *cobra.Command, args []string) {
			//TODO Prepare all CLI flags & configs here and then start call Run()

			dispatcher.Run()
		},
	}

	//cmd.PersistentFlags().

	return cmd
}
