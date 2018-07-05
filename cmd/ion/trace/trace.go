package trace

import (
	"fmt"
	"github.com/lawrencegripper/ion/cmd/ion/root"
	"github.com/lawrencegripper/ion/internal/pkg/management/trace"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"time"
)

//Client A shared GRPC module server client
var Client trace.TraceServiceClient
var managementEndpoint string
var timeoutSec int

var traceCmd = &cobra.Command{
	Use:               "trace",
	Short:             "trace gives you tools to view information about the execution of an item or module",
	PersistentPreRunE: Setup,
}

// Setup is called before Run and is used to setup any
// persistent components needed by sub commands.
func Setup(cmd *cobra.Command, args []string) error {
	if cmd.HasSubCommands() {
		return nil
	}

	fmt.Printf("using management endpoint %s\n", managementEndpoint)

	// Initialize a global GRPC connection to the management server
	conn, err := grpc.Dial(managementEndpoint,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(time.Duration(timeoutSec)*time.Second))

	if err != nil {
		return fmt.Errorf("failed to connect to server %s: %+v", managementEndpoint, err)
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

	// Local flags for the root command
	traceCmd.PersistentFlags().StringVar(&managementEndpoint, "endpoint", "localhost:9000", "management server endpoint")
	traceCmd.PersistentFlags().IntVar(&timeoutSec, "timeout", 30, "timeout in seconds for cli to connect to management server")
}
