package root

import (
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:     "ion",
	Short:   "ion cli lets you easily work with the ion system",
	PreRunE: Setup,
	RunE:    Root,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Root will run when no sub command is provided
func Root(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}

// Setup runs before the commands to initialise any global data
func Setup(cmd *cobra.Command, args []string) error {
	return nil
}

func init() {
	cobra.OnInitialize(initConfig)

	// Persistent flags shared by all subcommands
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ioncli.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".ioncli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".ioncli")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("using config file:", viper.ConfigFileUsed())
	}
}
