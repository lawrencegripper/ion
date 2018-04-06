package cmd

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "./config.yaml", "config file path")
	rootCmd.PersistentFlags().Bool("viper", true, "Use Viper for configuration")
	viper.BindPFlag("useViper", rootCmd.PersistentFlags().Lookup("viper"))
}

func initConfig() {
	viper.SetConfigFile(cfgFile)

	if err := viper.ReadInConfig(); err != nil {
		log.Errorln("Can't read config:", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "frontapi",
	Short: "frontapi receive external requests and generate the first event",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
