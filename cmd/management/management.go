package main

import (
	"strings"

	"github.com/lawrencegripper/ion/internal/app/management"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var managementConfig = management.NewConfiguration()

// NewManagementCommand create the management command with its flags
func NewManagementCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "ion-management",
		Short: "ion-management to embed in the processing",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			managementConfig.LogLevel = viper.GetString("log-level")
			managementConfig.PrintConfig, _ = cmd.Flags().GetBool("printconfig")

			// Globally set configuration level
			switch strings.ToLower(managementConfig.LogLevel) {
			case "debug":
				log.SetLevel(log.DebugLevel)
			case "info":
				log.SetLevel(log.InfoLevel)
			case "warn":
				log.SetLevel(log.WarnLevel)
			case "error":
				log.SetLevel(log.ErrorLevel)
			default:
				log.SetLevel(log.WarnLevel)
			}

			return nil
		},
	}

	cmd.AddCommand(NewStartCommand())

	flags := cmd.PersistentFlags()

	flags.StringP("loglevel", "l", "warn", "Logging level, possible values {debug, info, warn, error}")
	flags.BoolP("printconfig", "P", false, "Set to print config on start")

	_ = viper.BindPFlag("log-level", flags.Lookup("loglevel"))
	_ = viper.BindPFlag("print-config", flags.Lookup("printconfig"))

	return cmd
}
