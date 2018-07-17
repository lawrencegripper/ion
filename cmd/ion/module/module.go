package module

import (
	"fmt"
	"github.com/lawrencegripper/ion/cmd/ion/root"
	"github.com/lawrencegripper/ion/internal/pkg/management/module"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"time"
)

//Client A shared GRPC module server client
var Client module.ModuleServiceClient
var managementEndpoint, certFile string
var timeoutSec int

// moduleCmd represents the module command
var moduleCmd = &cobra.Command{
	Use:               "module",
	Short:             "execute commands to manage ion modules",
	PersistentPreRunE: Setup,
	Run:               Module,
}

// Module prints help
func Module(cmd *cobra.Command, args []string) {
	cmd.Help() // nolint: errcheck
}

// Setup is called before Run and is used to setup any
// persistent components needed by sub commands.
func Setup(cmd *cobra.Command, args []string) error {

	if cmd.HasSubCommands() {
		return nil
	}

	var options []grpc.DialOption

	if certFile != "" {
		creds, err := credentials.NewClientTLSFromFile(certFile, "")
		if err != nil {
			return fmt.Errorf("failed to load TLS configuraton from certificate file '%s': %+v", certFile, err)
		}
		options = append(options, grpc.WithTransportCredentials(creds))
	} else {
		options = append(options, grpc.WithInsecure())
	}

	options = append(options, grpc.WithBlock())
	options = append(options, grpc.WithTimeout(time.Duration(timeoutSec)*time.Second))

	// Initialize a global GRPC connection to the management server
	conn, err := grpc.Dial(managementEndpoint, options...)
	if err != nil {
		return fmt.Errorf("failed to connect to server %s: %+v", managementEndpoint, err)
	}
	Client = module.NewModuleServiceClient(conn)
	return nil
}

// Register adds to root command
func Register() {

	// Add module sub commands
	moduleCmd.AddCommand(createCmd)
	moduleCmd.AddCommand(deleteCmd)
	moduleCmd.AddCommand(listCmd)
	moduleCmd.AddCommand(getCmd)

	// Add module to root command
	root.RootCmd.AddCommand(moduleCmd)
}

func init() {

	// Local flags for the root command
	moduleCmd.PersistentFlags().StringVar(&managementEndpoint, "endpoint", "localhost:9000", "management server endpoint")
	moduleCmd.PersistentFlags().IntVar(&timeoutSec, "timeout", 30, "timeout in seconds for cli to connect to management server")
	moduleCmd.PersistentFlags().StringVar(&certFile, "certfile", "", "PEM formatted certificate file for mutual authentication")
}
