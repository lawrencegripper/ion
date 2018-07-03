package event

import (
	// "github.com/lawrencegripper/ion/internal/pkg/servicebus"
	"github.com/spf13/cobra"
)

var eventName string
var serviceBusConnectionString string

// createCmd represents the fire command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an event in the system",
	Run: func(cmd *cobra.Command, args []string) {
		//todo: add logic
	},
}

func init() {
	createCmd.Flags().StringVarP(&eventName, "event-name", "e", "test_event", "Name of the event to create (default: test_event)")
	createCmd.Flags().StringVarP(&serviceBusConnectionString, "amqp-connection-string", "a", "", "AMQP connection string ")
	createCmd.MarkFlagRequired("amqp-connection-string") //nolint: errcheck
}
