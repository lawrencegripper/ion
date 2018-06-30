package dev

import (
	"fmt"
	"github.com/lawrencegripper/ion/internal/app/handler"
	"github.com/lawrencegripper/ion/internal/app/handler/constants"
	"github.com/lawrencegripper/ion/internal/app/handler/development"
	"github.com/lawrencegripper/ion/internal/pkg/common"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

type prepareOptions struct {
	baseDir       string
	name          string
	eventID       string
	parentEventID string
	correlationID string
	eventTypes    []string
}

var prepareOpts prepareOptions

// prepareCmd represents the fire command
var prepareCmd = &cobra.Command{
	Use:   "prepare",
	Short: "initialise an origin module",
	RunE:  Prepare,
}

// Prepare a new origin module used to bootstrap a development flow
func Prepare(cmd *cobra.Command, args []string) error {

	config := handler.NewConfiguration()
	config.Action = constants.Prepare
	config.BaseDir = prepareOpts.baseDir
	config.Context = &common.Context{
		Name:          prepareOpts.name,
		EventID:       prepareOpts.eventID,
		CorrelationID: prepareOpts.correlationID,
		ParentEventID: prepareOpts.parentEventID,
	}
	config.ValidEventTypes = strings.Join(prepareOpts.eventTypes, ",")
	config.DevelopmentConfiguration = &development.Configuration{
		BaseDir: ionModulesDir,
		Enabled: true,
	}

	handler.Run(config)

	var inputBlobDir string
	err := filepath.Walk(ionModulesDir, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, filepath.FromSlash(filepath.Join(prepareOpts.parentEventID, "_blobs"))) {
			if !info.IsDir() {
				return nil // ignore files in the directory
			}
			inputBlobDir = path
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error reading parent module's blobs directory in %s: %+v", ionModulesDir, err)
	}
	moduleInputBlobDir := filepath.FromSlash(filepath.Join(prepareOpts.baseDir, "in", "data"))
	_ = os.RemoveAll(moduleInputBlobDir)
	if err := copyDir(inputBlobDir, moduleInputBlobDir); err != nil {
		return fmt.Errorf("error copying input blob data from %s to %s: %+v", inputBlobDir, moduleInputBlobDir, err)
	}

	fmt.Printf("module prepared, now run/debug your module. Set environment variable $HANDLER_BASE_DIR to '%s'\n", prepareOpts.baseDir)

	return nil
}

func init() {

	// Local flags to the prepare command
	prepareCmd.Flags().StringVar(&prepareOpts.baseDir, "base-dir", "ion", "base directory to run your module")
	prepareCmd.Flags().StringVar(&prepareOpts.name, "name", "", "module name")
	prepareCmd.Flags().StringSliceVar(&prepareOpts.eventTypes, "event-types", []string{}, "comma delimited string of valid event types for this module to publish")
	prepareCmd.Flags().StringVar(&prepareOpts.eventID, "event-id", "", "module's event id")
	prepareCmd.Flags().StringVar(&prepareOpts.correlationID, "correlation-id", "dev", "module's correlation id")
	prepareCmd.Flags().StringVar(&prepareOpts.parentEventID, "parent-event-id", "", "module's parent event id")

	// Mark required flags
	prepareCmd.MarkFlagRequired("name")            //nolint: errcheck
	prepareCmd.MarkFlagRequired("event-types")     //nolint: errcheck
	prepareCmd.MarkFlagRequired("event-id")        //nolint: errcheck
	prepareCmd.MarkFlagRequired("parent-event-id") //nolint: errcheck
}
