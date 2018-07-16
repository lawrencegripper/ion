package module

import (
	"fmt"

	"context"
	"github.com/lawrencegripper/ion/internal/pkg/management/module"
	"github.com/spf13/cobra"
)

type deleteOptions struct {
	name string
}

var deleteOpts deleteOptions

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete a module from ion",
	RunE:  Delete,
}

// Delete an ion module
func Delete(cmd *cobra.Command, args []string) error {

	deleteRequest := &module.ModuleDeleteRequest{
		Name: deleteOpts.name,
	}

	fmt.Println("deleting module")
	_, err := Client.Delete(context.Background(), deleteRequest)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("failed to delete module: %+v", err))
	}
	fmt.Printf("deleted module %s\n", deleteOpts.name)
	return nil
}

func init() {

	// Local flags for the delete command
	deleteCmd.Flags().StringVarP(&deleteOpts.name, "name", "n", "", "the module name")

	// Mark required flags
	deleteCmd.MarkFlagRequired("name") //nolint: errcheck
}
