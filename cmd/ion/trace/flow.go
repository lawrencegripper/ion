package trace

import (
	"context"
	"fmt"
	"github.com/lawrencegripper/ion/internal/pkg/management/trace"
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
	response, err := Client.GetFlow(context.Background(), &trace.GetFlowRequest{
		CorrelationID: correlationID,
	})
	if err != nil {
		return err
	}

	fmt.Println(response.FlowJSON)
	return nil
}

func init() {

	// Local flags for the create command
	flowCmd.Flags().StringVarP(&correlationID, "correlationid", "c", "", "provide a correlationID of an item")

	// Mark requried flags
	flowCmd.MarkFlagRequired("correlationid") //nolint: errcheck
}
