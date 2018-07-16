package module

import (
	"fmt"

	"context"
	"encoding/json"
	"github.com/lawrencegripper/ion/internal/pkg/management/module"
	"github.com/spf13/cobra"
)

type getOptions struct {
	name string
}

var getOpts getOptions

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get a module from ion",
	RunE:  Get,
}

// Get an ion module
func Get(cmd *cobra.Command, args []string) error {

	getRequest := &module.ModuleGetRequest{
		Name: getOpts.name,
	}
	getResponse, err := Client.Get(context.Background(), getRequest)
	if err != nil {
		return fmt.Errorf("failed to get module %s with error %+v", getOpts.name, err)
	}
	b, err := json.Marshal(getResponse)
	if err != nil {
		return fmt.Errorf("error parsing response from server %+v", err)
	}
	fmt.Println(string(b))
	return nil
}

func init() {

	// Local flags for the get command
	getCmd.Flags().StringVarP(&getOpts.name, "name", "n", "", "the module name")

	// Mark required flags
	getCmd.MarkFlagRequired("name") //nolint: errcheck
}
