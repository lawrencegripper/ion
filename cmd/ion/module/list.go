package module

import (
	"fmt"

	"context"
	"github.com/lawrencegripper/ion/internal/pkg/management/module"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list currently running modules in ion",
	RunE:  List,
}

// List all the current ion modules managed by this client
func List(cmd *cobra.Command, args []string) error {
	listRequest := &module.ModuleListRequest{}

	fmt.Println("listing all modules")
	listResponse, err := Client.List(context.Background(), listRequest)
	if err != nil {
		return fmt.Errorf("failed to list module: %+v", err)
	}
	for _, name := range listResponse.Names {
		fmt.Printf("%s\n", name)
	}
	return nil
}

func init() {}
