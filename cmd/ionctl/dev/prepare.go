package dev

import (
	"fmt"
	// "github.com/lawrencegripper/ion/internal/pkg/servicebus"
	"github.com/spf13/cobra"
)

// prepareCmd represents the fire command
var prepareCmd = &cobra.Command{
	Use:   "prepare",
	Short: "initialise an origin module",
	RunE:  Prepare,
}

// Prepare a new origin module used to bootstrap a development flow
func Prepare(cmd *cobra.Command, args []string) error {
	fmt.Println("prepare module")

	return nil
}

func init() {
}
