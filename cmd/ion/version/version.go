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

// Version prints the current version information
func Version(cmd *cobra.Command, args []string) {
	clientVer := GetClientVersion()

	//TODO: Support output formats
	fmt.Printf("%+v\n", clientVer)
}

// Register adds to root command
func Register() {
	// Add version command to root
	root.RootCmd.AddCommand(versionCmd)
}
