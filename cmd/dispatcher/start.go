package main

import (
	"errors"

	"github.com/spf13/viper"

	"github.com/lawrencegripper/ion/internal/app/dispatcher"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// NewCmdStart return cobra.Command to run ion-disptacher start command
func NewCmdStart() *cobra.Command {
	cmdStart := &cobra.Command{
		Use:   "start",
		Short: "Instanciate the dispatcher to process events",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Fill config with command settings
			cfg.ClientID = viper.GetString("clientid")
			cfg.ClientSecret = viper.GetString("clientsecret")
			cfg.SubscriptionID = viper.GetString("subscriptionid")
			cfg.TenantID = viper.GetString("tenantid")

			printConfig()

			if cfg.ClientID == "" || cfg.ClientSecret == "" || cfg.TenantID == "" || cfg.SubscriptionID == "" {
				return errConfigurationMissing
			}
			if cfg.Job == nil {
				return errors.New("Job config can't be nil")
			}
			if cfg.Sidecar == nil {
				return errors.New("Sidecar config can't be nil")
			}
			//TODO: validate sidecar config

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			log.Infoln("Starting dispatcher")

			dispatcher.Run(&cfg)
		},
	}

	// Add 'dispatcher start' flags
	cmdStart.PersistentFlags().String("clientid", "", "ClientID of Service Principal for Azure access")
	cmdStart.PersistentFlags().String("clientsecret", "", "Client Secrete of Service Principal for Azure access")
	cmdStart.PersistentFlags().String("subscriptionid", "", "SubscriptionID for Azure")
	cmdStart.PersistentFlags().String("tenantid", "", "TentantID for Azure")

	// Mark required flags (won't mark required setting, onyl CLI flag)
	//cmdStart.MarkPersistentFlagRequired("")

	// Bing flags and config file values
	viper.BindPFlag("clientid", cmdStart.PersistentFlags().Lookup("clientid"))
	viper.BindPFlag("clientsecret", cmdStart.PersistentFlags().Lookup("clientsecret"))
	viper.BindPFlag("subscriptionid", cmdStart.PersistentFlags().Lookup("subscriptionid"))
	viper.BindPFlag("tenantid", cmdStart.PersistentFlags().Lookup("tenantid"))

	return cmdStart
}
