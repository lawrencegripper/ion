package trace

import (
	"github.com/lawrencegripper/ion/cmd/ion/root"
	"github.com/lawrencegripper/ion/internal/pkg/management/trace"
	"github.com/spf13/cobra"
)

//Client A shared GRPC module server client
var Client trace.TraceServiceClient

var traceCmd = &cobra.Command{
	Use:               "trace",
	Short:             "trace gives you tools to view information about the execution of an item or module",
	PersistentPreRunE: Setup,
}

// Setup is called before Run and is used to setup any
// persistent components needed by sub commands.
func Setup(cmd *cobra.Command, args []string) error {
	conn, err := root.GetManagementConnection()
	if err != nil {
		return err
	}
	Client = trace.NewTraceServiceClient(conn)
	return nil
}

// Register adds to root command
func Register() {
	// Add module sub commands
	traceCmd.AddCommand(flowCmd)

	// Add module to root command
	root.RootCmd.AddCommand(traceCmd)
}

func init() {
}
