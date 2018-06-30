package dev

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

type listModulesOptions struct {
	rootID string
}

var listModulesOpts listModulesOptions

// listModulesCmd represents the fire command
var listModulesCmd = &cobra.Command{
	Use:   "list",
	Short: "list a development module's events",
	RunE:  ListModules,
}

// List a module's events
func ListModules(cmd *cobra.Command, args []string) error {

	var files []string

	err := filepath.Walk(ionModulesDir, func(path string, info os.FileInfo, err error) error {
		if path == ionModulesDir {
			return nil // skip root
		}
		if !info.IsDir() {
			return nil // skip files
		}
		if filepath.Base(path)[0] == '_' {
			return nil // ignore data files
		}
		if listModulesOpts.rootID == "" || strings.Contains(path, listModulesOpts.rootID) {
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error finding module directories in %s: %+v", ionModulesDir, err)
	}
	if len(files) == 0 {
		fmt.Println("no modules available")
		return nil
	}

	for _, file := range files {
		file := strings.Replace(file, ionModulesDir, "", -1)
		fmt.Println(file)
	}

	return nil
}

func init() {
	// Local flags to the prepare command
	listModulesCmd.Flags().StringVar(&listModulesOpts.rootID, "root-event-id", "", "only list children of the given event id")
}
