package module

import (
	"fmt"

	"github.com/lawrencegripper/ion/cmd/ionctl/root"
	"github.com/spf13/cobra"
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
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("module called")
	},
}

// Register adds to root command
func Register() {
	root.RootCmd.AddCommand(moduleCmd)
}
