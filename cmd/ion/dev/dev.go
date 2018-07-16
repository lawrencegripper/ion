package dev

import (
	"fmt"
	"github.com/lawrencegripper/ion/cmd/ion/root"
	"path/filepath"

	"github.com/spf13/cobra"
	"os"
	"runtime"
)

var (
	ionHomeDir    = filepath.FromSlash(filepath.Join(UserHomeDir(), ".ion"))
	ionStoreDir   = filepath.FromSlash(filepath.Join(ionHomeDir, "store"))
	ionModulesDir = filepath.FromSlash(filepath.Join(ionStoreDir, "modules"))
)

// devCmd represents the events command
var devCmd = &cobra.Command{
	Use:               "dev",
	Short:             "execute development operations",
	PersistentPreRunE: Setup,
	RunE:              Dev,
}

// Setup runs before Dev and makes sure the ion store is created
func Setup(cmd *cobra.Command, args []string) error {
	_, err := os.Stat(ionStoreDir)
	if err != nil {
		if err := os.MkdirAll(ionStoreDir, 0777); err != nil {
			return fmt.Errorf("error creating ion store at location %s: %+v", ionStoreDir, err)
		}
	}
	return nil
}

// Dev prints the available sub commands
func Dev(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}

// Register adds to root command
func Register() {

	// Add module sub commands
	moduleCmd.AddCommand(initCmd)
	moduleCmd.AddCommand(commitCmd)
	moduleCmd.AddCommand(prepareCmd)
	moduleCmd.AddCommand(listModulesCmd)
	moduleCmd.AddCommand(selectCmd)

	// Add event sub commands
	eventCmd.AddCommand(listEventsCmd)

	// Add event sub commands
	devCmd.AddCommand(moduleCmd)
	devCmd.AddCommand(eventCmd)

	// Add event command to root
	root.RootCmd.AddCommand(devCmd)
}

// UserHomeDir returns the user's home directory
func UserHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}
