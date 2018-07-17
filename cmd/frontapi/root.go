package main

import (
	"fmt"
	"os"

	"github.com/lawrencegripper/ion/internal/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var cfg = types.Configuration{
	Job: &types.JobConfig{},
	Handler: &types.HandlerConfig{
		MongoDBDocumentStorageProvider: &types.MongoDBConfig{},
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "../../configs/frontapi.yaml", "Config file path")
	rootCmd.PersistentFlags().Bool("viper", true, "Use Viper for configuration")
	_ = viper.BindPFlag("useViper", rootCmd.PersistentFlags().Lookup("viper"))
}

var rootCmd = &cobra.Command{
	Use:   "frontapi",
	Short: "frontapi receive external requests and generate the first event",
}

func execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
