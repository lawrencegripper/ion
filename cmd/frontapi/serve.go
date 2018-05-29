package main

import (
	"os"

	log "github.com/Sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/lawrencegripper/ion/internal/app/frontapi"
)

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.PersistentFlags().Int("port", 8080, "Listenning port")

	viper.BindPFlag("port", serveCmd.PersistentFlags().Lookup("port"))
	viper.BindEnv("servicebus_namespace")
	viper.BindEnv("servicebus_topic")
	viper.BindEnv("servicebus_saspolicy")
	viper.BindEnv("servicebus_accesskey")

	viper.SetEnvPrefix("frontapi")
	viper.AutomaticEnv()
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve the HTTP handlers of frontapi",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		viper.SetConfigFile(cfgFile)
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err != nil {
			log.WithError(err).Errorln("Can't read config")
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		frontapi.Run(&cfg, viper.GetInt("port"))
	},
}
