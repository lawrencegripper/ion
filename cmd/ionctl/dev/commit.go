package dev

import (
	"fmt"
	// "github.com/lawrencegripper/ion/internal/pkg/servicebus"
	"github.com/spf13/cobra"
)

// initCmd represents the fire command
var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "commit a module's state to the development dataplane",
	RunE:  Commit,
}

// Commit a new origin module used to bootstrap a development flow
func Commit(cmd *cobra.Command, args []string) error {
	fmt.Println("commit module")

	return nil
}

func init() {
}
