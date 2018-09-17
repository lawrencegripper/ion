package module

import (
	"github.com/lawrencegripper/ion/cmd/ion/root"
	"github.com/lawrencegripper/ion/internal/pkg/management/module"
	"github.com/spf13/cobra"
)

// Client to be used by any subcommands to talk to the module service
var Client module.ModuleServiceClient

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
	conn, err := root.GetManagementConnection()
	if err != nil {
		return err
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
}
