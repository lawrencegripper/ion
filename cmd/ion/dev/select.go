package dev

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

type selectOptions struct {
	moduleID string
}

var selectOpts selectOptions

// selectCmd represents the fire command
var selectCmd = &cobra.Command{
	Use:   "select",
	Short: "select a development module",
	RunE:  Select,
}

// Select a module
func Select(cmd *cobra.Command, args []string) error {

	err := filepath.Walk(ionModulesDir, func(path string, info os.FileInfo, err error) error {
		if path == ionModulesDir {
			return nil // skip root
		}
		if !info.IsDir() {
			return nil // skip files
		}
		if filepath.Base(path) == selectOpts.moduleID {
			ionModuleDir := "ion-module"
			_, err := os.Lstat(ionModuleDir)
			if err == nil {
				_ = os.Remove(ionModuleDir)
			}
			abs, err := filepath.Abs(path)
			if err != nil {
				return fmt.Errorf("error creating absolute path for %s: %+v", path, err)
			}
			if err := os.Symlink(abs, ionModuleDir); err != nil {
				return fmt.Errorf("error creating symblink between %s and %s", abs, ionModuleDir)
			}

			fmt.Printf("module selected and available in directory %s\n", ionModuleDir)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error selecting module's directory: %+v err: %+v", ionModulesDir, err)
	}

	return nil
}

func init() {
	// Local flags to the prepare command
	selectCmd.Flags().StringVar(&selectOpts.moduleID, "module-id", "", "the ion development module you wish to select")

	// Mark flags required
	selectCmd.MarkFlagRequired("module-id") //nolint: errcheck
}
