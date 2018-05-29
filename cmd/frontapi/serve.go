package main

import (
	"errors"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/lawrencegripper/ion/internal/app/frontapi"
)

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.PersistentFlags().Int("port", 8080, "Listenning port")

	// Add 'dispatcher' flags
	serveCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "../../configs/dispatcher.yaml", "Config file path")
	serveCmd.PersistentFlags().StringP("loglevel", "l", "warn", "Log level (debug|info|warn|error)")
	serveCmd.PersistentFlags().String("modulename", "", "Name of the module")
	serveCmd.PersistentFlags().String("subscribestoevent", "", "Event this modules subscribes to")
	serveCmd.PersistentFlags().String("eventspublished", "", "Events this modules can publish")
	serveCmd.PersistentFlags().String("servicebusnamespace", "", "Namespace to use for ServiceBus")
	serveCmd.PersistentFlags().String("resourcegroup", "", "Azure ResourceGroup to use")
	serveCmd.PersistentFlags().Bool("logsensitiveconfig", false, "Print out sensitive config when logging")
	serveCmd.PersistentFlags().String("moduleconfigpath", "", "Path to environment variables file for module")
	serveCmd.PersistentFlags().BoolP("printconfig", "P", false, "Print out config when starting")
	// job.*
	serveCmd.PersistentFlags().Int("job.retrycount", 0, "Max number of times a job can be retried")

	viper.BindPFlag("port", serveCmd.PersistentFlags().Lookup("port"))

	viper.SetEnvPrefix("frontapi")
	viper.AutomaticEnv()
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve the HTTP handlers of frontapi",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		viper.SetConfigFile(cfgFile)
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err != nil {
			log.WithError(err).Errorln("Can't read config")
			os.Exit(1)
		}
		viper.AutomaticEnv()

		// Fill config with global settings
		cfg.LogLevel = viper.GetString("loglevel")
		cfg.ModuleName = viper.GetString("modulename")
		cfg.SubscribesToEvent = viper.GetString("subscribestoevent")
		cfg.EventsPublished = viper.GetString("eventspublished")
		cfg.ServiceBusNamespace = viper.GetString("servicebusnamespace")
		cfg.ResourceGroup = viper.GetString("resourcegroup")
		cfg.PrintConfig = viper.GetBool("printconfig")

		// job.*
		cfg.Job.RetryCount = viper.GetInt("job.retrycount")

		// Globally set configuration level
		switch strings.ToLower(cfg.LogLevel) {
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

		hostName, err := os.Hostname()
		if err != nil {
			return errors.New("Unable to automatically set instanceid to hostname")
		}
		cfg.Hostname = hostName

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		frontapi.Run(&cfg, viper.GetInt("port"))
	},
}
