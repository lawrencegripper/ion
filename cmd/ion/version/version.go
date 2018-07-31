package version

import (
	"fmt"
	"github.com/lawrencegripper/ion/cmd/ion/root"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show ion client version",
	Run:   Version,
}

// Setup is called before Run and is used to setup any
// persistent components needed by sub commands.
func Version(cmd *cobra.Command, args []string) {
	ver := GetVersion()

	//TODO: Support output formats
	fmt.Printf("%+v\n", ver)
}

// Register adds to root command
func Register() {
	// Add version command to root
	root.RootCmd.AddCommand(versionCmd)
}
