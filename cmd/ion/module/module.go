package module

import (
	"fmt"
	"github.com/lawrencegripper/ion/cmd/ion/root"
	"github.com/lawrencegripper/ion/internal/pkg/management/module"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// moduleCmd represents the module command
var moduleCmd = &cobra.Command{
	Use:   "module",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
}

// Register adds to root command
func Register() {
	root.RootCmd.AddCommand(moduleCmd)
}

func getConnection() module.ModuleServiceClient {
	conn, err := grpc.Dial(root.ManagementAPIEndpoint, grpc.WithInsecure())
	if err != nil {
		panic(fmt.Sprintf("failed to dial server: %+v", err))
	}
	return module.NewModuleServiceClient(conn)
}
