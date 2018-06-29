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

	flags := serveCmd.PersistentFlags()
	// Add 'dispatcher' flags
	flags.StringVarP(&cfgFile, "config", "c", "../../configs/dispatcher.yaml", "Config file path")
	flags.StringP("loglevel", "l", "warn", "Log level (debug|info|warn|error)")
	flags.String("modulename", "", "Name of the module")
	flags.String("subscribestoevent", "", "Event this modules subscribes to")
	flags.String("eventspublished", "", "Events this modules can publish")
	flags.String("servicebusnamespace", "", "Namespace to use for ServiceBus")
	flags.String("resourcegroup", "", "Azure ResourceGroup to use")
	flags.Bool("logsensitiveconfig", false, "Print out sensitive config when logging")
	flags.String("moduleconfigpath", "", "Path to environment variables file for module")
	flags.BoolP("printconfig", "P", false, "Print out config when starting")

	// job.*
	flags.Int("job.retrycount", 0, "Max number of times a job can be retried")

	// document store flags
	flags.String("mongodb-name", "", "MongoDB Name")
	flags.String("mongodb-collection", "", "MongoDB Database Collection")
	flags.String("mongodb-username", "", "MongoDB server username")
	flags.String("mongodb-password", "", "MongoDB server password")
	flags.Int("mongodb-port", 27017, "MongoDB server port")

	_ = viper.BindPFlag("port", serveCmd.PersistentFlags().Lookup("port"))

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

		cfg.Handler.MongoDBDocumentStorageProvider.Name = viper.GetString("mongodb-name")
		cfg.Handler.MongoDBDocumentStorageProvider.Collection = viper.GetString("mongodb-collection")
		cfg.Handler.MongoDBDocumentStorageProvider.Password = viper.GetString("mongo-password")
		cfg.Handler.MongoDBDocumentStorageProvider.Port = viper.GetInt("mongo-port")

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
