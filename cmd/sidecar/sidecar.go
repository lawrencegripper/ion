package main

import (
	"strings"

	"github.com/lawrencegripper/ion/internal/app/sidecar"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var sidecarConfig = sidecar.NewConfiguration()

var sidecarCmdConfig = viper.New()

// NewSidecarCommand create the sidecar command with its flags
func NewSidecarCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "ion-sidecar",
		Short: "ion-sidecar to embed in the processing",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			sidecarConfig.LogFile = sidecarCmdConfig.GetString("log-file")
			sidecarConfig.LogLevel = sidecarCmdConfig.GetString("log-level")
			sidecarConfig.Development = sidecarCmdConfig.GetBool("development")
			sidecarConfig.PrintConfig, _ = cmd.Flags().GetBool("printconfig")

			// Globally set configuration level
			switch strings.ToLower(sidecarConfig.LogLevel) {
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
	cmd.AddCommand(NewVersionCommand())

	flags := cmd.PersistentFlags()

	flags.StringP("logfile", "L", "", "File to log output to")
	flags.StringP("loglevel", "l", "warn", "Logging level, possible values {debug, info, warn, error}")
	flags.BoolP("development", "d", false, "A flag to enable development features")
	flags.BoolP("printconfig", "P", false, "Set to print config on start")

	sidecarCmdConfig.BindPFlag("log-file", flags.Lookup("logfile"))
	sidecarCmdConfig.BindPFlag("log-level", flags.Lookup("loglevel"))
	sidecarCmdConfig.BindPFlag("development", flags.Lookup("development"))

	return cmd
}
