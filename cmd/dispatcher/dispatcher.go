package main

import (
	"os"

	"github.com/lawrencegripper/ion/internal/pkg/tools"
	"github.com/lawrencegripper/ion/internal/pkg/types"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfg types.Configuration

func init() {
	//cobra.OnInitialize(initConfig)

	//TODO We should think at a nicier place for storing configuration files, like $HOME/.config/$SERVICE/...
	//rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "./configs/config.yaml", "Config file path")
	//rootCmd.PersistentFlags().StringVarP("loglevel", "l", "warn", "Log level")

	//viper.BindPFlag("googlesearchservice.verbose", rootCmd.PersistentFlags().Lookup("verbose"))

	//viper.BindEnv("googlesearchservice.port", "GOOGLESEARCHSERVICE_PORT")
}

func initConfig() {
}

// NewDispatcherCommand return cobra.Command to run ion-disptacher command
func NewDispatcherCommand() *cobra.Command {
	dispatcherCmd := &cobra.Command{
		Use:   "ion-dispatcher",
		Short: "ion-dispatcher: ...",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cfg.LogLevel = viper.GetString("loglevel")
			cfg.PrintConfig = viper.GetBool("printconfig")
			cfg.LogSensitiveConfig = viper.GetBool("logsensitiveconfig")

			if cfg.PrintConfig {
				if cfg.LogSensitiveConfig {
					log.Infoln(tools.PrettyPrintStruct(cfg))
				} else {
					log.Infoln(tools.PrettyPrintStruct(types.RedactConfigSecrets(&cfg)))
				}
			}
		},
	}

	// Add dispatcher flags
	dispatcherCmd.PersistentFlags().String("hostname", "", "")
	dispatcherCmd.PersistentFlags().StringP("loglevel", "l", "warn", "Log level")
	dispatcherCmd.PersistentFlags().String("modulename", "", "Name of the module")
	dispatcherCmd.PersistentFlags().String("subscribestoevent", "", "Event this modules subscribes to")
	dispatcherCmd.PersistentFlags().String("eventspublished", "", "Events this modules can publish")
	dispatcherCmd.PersistentFlags().String("servicebusnamespace", "", "Namespace to use for ServiceBus")
	dispatcherCmd.PersistentFlags().String("resourcegroup", "", "Azure ResourceGroup to use")
	dispatcherCmd.PersistentFlags().String("subscriptionid", "", "SubscriptionID for Azure")
	dispatcherCmd.PersistentFlags().String("clientid", "", "ClientID of Service Principal for Azure access")
	dispatcherCmd.PersistentFlags().String("clientsecret", "", "Client Secrete of Service Principal for Azure access")
	dispatcherCmd.PersistentFlags().String("tenantid", "", "TentantID for Azure")
	dispatcherCmd.PersistentFlags().Bool("logsensitiveconfig", false, "Print out sensitive config when logging")
	dispatcherCmd.PersistentFlags().String("moduleconfigpath", "", "Path to environment variables file for module")
	dispatcherCmd.PersistentFlags().Bool("printconfig", false, "Print out sensitive config when logging")

	// Read config file
	var cfgFile string
	dispatcherCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "../../configs/dispatcher.yaml", "Config file path")
	viper.SetConfigFile(cfgFile)
	if err := viper.ReadInConfig(); err != nil {
		log.Errorln("Can't read config:", err)
		os.Exit(1)
	}

	// Bing flags and config files
	viper.BindPFlag("loglevel", dispatcherCmd.PersistentFlags().Lookup("loglevel"))
	viper.BindPFlag("logsensitiveconfig", dispatcherCmd.PersistentFlags().Lookup("logsensitiveconfig"))
	viper.BindPFlag("printconfig", dispatcherCmd.PersistentFlags().Lookup("printconfig"))

	// Add sub-commands
	dispatcherCmd.AddCommand(NewCmdStart())

	return dispatcherCmd
}
