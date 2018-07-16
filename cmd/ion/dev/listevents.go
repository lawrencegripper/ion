package dev

import (
	"encoding/json"
	"fmt"
	"github.com/lawrencegripper/ion/internal/pkg/common"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type listEventsOptions struct {
	moduleID string
}

var listEventsOpts listEventsOptions

// listEventsCmd represents the fire command
var listEventsCmd = &cobra.Command{
	Use:   "list",
	Short: "list a development module's events",
	RunE:  ListEvents,
}

//ListEvents List a module's events
func ListEvents(cmd *cobra.Command, args []string) error {

	var files []string

	err := filepath.Walk(ionModulesDir, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, filepath.FromSlash(filepath.Join(listEventsOpts.moduleID, "_events"))) {
			if info.IsDir() {
				return nil // ignore events directory
			}
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("error reading module's event directory in %s: %+v", ionModulesDir, err)
	}
	if len(files) == 0 {
		fmt.Println("module raised no events")
		return nil
	}

	for _, file := range files {
		b, err := ioutil.ReadFile(file)
		if err != nil {
			return fmt.Errorf("error reading module event %s: %+v", file, err)
		}
		var event common.Event
		if err = json.Unmarshal(b, &event); err != nil {
			return fmt.Errorf("error parsing module event %s: %+v", file, err)
		}
		fmt.Println(string(b))
	}

	return nil
}

func init() {
	// Local flags to the prepare command
	listEventsCmd.Flags().StringVar(&listEventsOpts.moduleID, "module-id", "", "the id of the module you want to list the events for")

	// Mark required flags
	listEventsCmd.MarkFlagRequired("module-id") //nolint: errcheck
}
