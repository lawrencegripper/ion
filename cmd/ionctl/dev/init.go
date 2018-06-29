package dev

import (
	"fmt"
	// "github.com/lawrencegripper/ion/internal/pkg/servicebus"
	"github.com/spf13/cobra"
)

// initCmd represents the fire command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "initialise an origin module",
	RunE:  Init,
}

// Init a new origin module used to bootstrap a development flow
func Init(cmd *cobra.Command, args []string) error {
	fmt.Println("init module")

	return nil
}

func init() {
}
