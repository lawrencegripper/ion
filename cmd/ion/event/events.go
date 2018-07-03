package event

import (
	"github.com/lawrencegripper/ion/cmd/ion/root"

	"github.com/spf13/cobra"
)

// eventsCmd represents the events command
var eventCmd = &cobra.Command{
	Use:   "event",
	Short: "Manage events in the system",
}

// Register adds to root command
func Register() {
	root.RootCmd.AddCommand(eventCmd)
	eventCmd.AddCommand(createCmd)
}
