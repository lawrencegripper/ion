package trace

import (
	"github.com/spf13/cobra"
)

var correlationID string

// flowCmd represents the create command
var flowCmd = &cobra.Command{
	Use:   "flow",
	Short: "follow the flow of an item through the modules via it's correlationID",
	RunE:  flow,
}

// flow a new ion module
func flow(cmd *cobra.Command, args []string) error {
	return nil
}

func init() {

	// Local flags for the create command
	flowCmd.Flags().StringVarP(&correlationID, "correlationid", "c", "", "provide a correlationID of an item")

	// Mark requried flags
	flowCmd.MarkFlagRequired("correlationid") //nolint: errcheck
}
