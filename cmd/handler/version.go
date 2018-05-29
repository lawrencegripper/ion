// nolint: errcheck
package main

import (
	"fmt"

	"github.com/lawrencegripper/ion/internal/app/handler"
	"github.com/spf13/cobra"
)

// NewVersionCommand print the handler version
func NewVersionCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "version",
		Short: "Get handler's version",
		Run:   func(cmd *cobra.Command, args []string) { fmt.Println(handler.Version) },
	}
	return cmd
}
