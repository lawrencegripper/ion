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

	//TODO We should think at a nicier place for storing configuration files, like $HOME/.config/$SERVICE/...
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "./configs/config.yaml", "Config file path")
}

func initConfig() {
	viper.SetConfigFile(cfgFile)

	if err := viper.ReadInConfig(); err != nil {
		log.Errorln("Can't read config:", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use: "ion",
	//Short: "",
	//Run: func(cmd *cobra.Command, args []string) {
	//},
}

//Execute launch the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
