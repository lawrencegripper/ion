package dev

import (
	"fmt"
	"github.com/lawrencegripper/ion/internal/app/handler"
	"github.com/lawrencegripper/ion/internal/app/handler/constants"
	"github.com/lawrencegripper/ion/internal/app/handler/development"
	"github.com/lawrencegripper/ion/internal/pkg/common"
	"github.com/spf13/cobra"
	"strings"
)

type commitOptions struct {
	baseDir       string
	name          string
	eventID       string
	parentEventID string
	correlationID string
	eventTypes    []string
}

var commitOpts commitOptions

// initCmd represents the fire command
var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "commit a module's state to the development dataplane",
	RunE:  Commit,
}

// Commit a new origin module used to bootstrap a development flow
func Commit(cmd *cobra.Command, args []string) error {
	config := handler.NewConfiguration()
	config.Action = constants.Commit
	config.BaseDir = commitOpts.baseDir
	config.Context = &common.Context{
		Name:          commitOpts.name,
		EventID:       commitOpts.eventID,
		CorrelationID: commitOpts.correlationID,
		ParentEventID: commitOpts.parentEventID,
	}
	config.ValidEventTypes = strings.Join(commitOpts.eventTypes, ",")
	config.DevelopmentConfiguration = &development.Configuration{
		BaseDir: ionModulesDir,
		Enabled: true,
	}

	handler.Run(config)

	fmt.Println("module has been committed")

	return nil
}

func init() {
	// Local flags to the prepare command
	commitCmd.Flags().StringVar(&commitOpts.baseDir, "base-dir", "ion", "base directory to run your module")
	commitCmd.Flags().StringVar(&commitOpts.name, "name", "", "module name")
	commitCmd.Flags().StringSliceVar(&commitOpts.eventTypes, "event-types", []string{}, "comma delimited string of valid event types for this module to publish")
	commitCmd.Flags().StringVar(&commitOpts.eventID, "event-id", "", "module's event id")
	commitCmd.Flags().StringVar(&commitOpts.correlationID, "correlation-id", "dev", "module's correlation id")
	commitCmd.Flags().StringVar(&commitOpts.parentEventID, "parent-event-id", "", "module's parent event id")

	// Mark required flags
	commitCmd.MarkFlagRequired("name")            //nolint: errcheck
	commitCmd.MarkFlagRequired("event-types")     //nolint: errcheck
	commitCmd.MarkFlagRequired("event-id")        //nolint: errcheck
	commitCmd.MarkFlagRequired("parent-event-id") //nolint: errcheck
}
