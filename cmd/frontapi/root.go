package main

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"

	"github.com/lawrencegripper/ion/internal/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var cfg = types.Configuration{
	Kubernetes: &types.KubernetesConfig{},
	Job:        &types.JobConfig{},
	Handler: &types.HandlerConfig{
		AzureBlobStorageProvider:       &types.AzureBlobConfig{},
		MongoDBDocumentStorageProvider: &types.MongoDBConfig{},
	},
	AzureBatch: &types.AzureBatchConfig{},
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "../../configs/frontapi.yaml", "Config file path")
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
