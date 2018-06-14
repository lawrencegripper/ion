package main

import (
	"fmt"

	"github.com/lawrencegripper/ion/internal/app/management"
	"github.com/lawrencegripper/ion/internal/pkg/tools"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// NewStartCommand create the start command with its flags
func NewStartCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "start",
		Short: "ion-management to embed in the processing",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Read config file
			viper.SetConfigFile(cfgFile)
			if err := viper.ReadInConfig(); err != nil {
				log.WithError(err).Warningln("Can't read config")
			}
			viper.AutomaticEnv()

			managementConfig.Port = viper.GetInt("management-port")
			managementConfig.Namespace = viper.GetString("namespace")
			managementConfig.DispatcherImage = viper.GetString("dispatcher-image-name")
			managementConfig.DispatcherImageTag = viper.GetString("dispatcher-image-tag")
			managementConfig.AzureClientID = viper.GetString("azure-client-id")
			managementConfig.AzureClientSecret = viper.GetString("azure-client-secret")
			managementConfig.AzureSubscriptionID = viper.GetString("azure-subscription-id")
			managementConfig.AzureTenantID = viper.GetString("azure-tenant-id")
			managementConfig.AzureResourceGroup = viper.GetString("azure-resource-group")
			managementConfig.AzureBatchPoolID = viper.GetString("azure-batch-pool-id")
			managementConfig.AzureBatchAccountName = viper.GetString("azure-batch-account-name")
			managementConfig.AzureBatchAccountLocation = viper.GetString("azure-batch-account-location")
			managementConfig.MongoDBName = viper.GetString("mongodb-name")
			managementConfig.MongoDBPort = viper.GetInt("mongodb-port")
			managementConfig.MongoDBCollection = viper.GetString("mongodb-collection")
			managementConfig.MongoDBPassword = viper.GetString("mongodb-password")
			managementConfig.AzureStorageAccountName = viper.GetString("azure-storage-account-name")
			managementConfig.AzureStorageAccountKey = viper.GetString("azure-storage-account-key")
			managementConfig.LogLevel = viper.GetString("loglevel")

			if managementConfig.DispatcherImage == "" {
				return fmt.Errorf("--dispatcher-image-name is required")
			}
			if managementConfig.DispatcherImageTag == "" {
				return fmt.Errorf("--dispatcher-image-tag is required")
			}
			if managementConfig.AzureClientID == "" {
				return fmt.Errorf("--azure-client-id is required")
			}
			if managementConfig.AzureClientSecret == "" {
				return fmt.Errorf("--azure-client-secret is required")
			}
			if managementConfig.AzureSubscriptionID == "" {
				return fmt.Errorf("--azure-subscription-id is required")
			}
			if managementConfig.AzureTenantID == "" {
				return fmt.Errorf("--azure-tenant-id is required")
			}
			if managementConfig.AzureResourceGroup == "" {
				return fmt.Errorf("--azure-resource-group is required")
			}
			// Should be optional?
			if managementConfig.AzureBatchPoolID == "" {
				return fmt.Errorf("--azure-batch-pool-id is required")
			}
			if managementConfig.AzureBatchAccountName == "" {
				return fmt.Errorf("--azure-batch-account-name is required")
			}
			if managementConfig.AzureBatchAccountLocation == "" {
				return fmt.Errorf("--azure-batch-account-location is required")
			}
			if managementConfig.MongoDBName == "" {
				return fmt.Errorf("--mongodb-name is required")
			}
			if managementConfig.MongoDBCollection == "" {
				return fmt.Errorf("--mongodb-collection is required")
			}
			if managementConfig.MongoDBPassword == "" {
				return fmt.Errorf("--mongodb-password is required")
			}
			if managementConfig.AzureStorageAccountName == "" {
				return fmt.Errorf("--azure-storage-account-name is required")
			}
			if managementConfig.AzureStorageAccountKey == "" {
				return fmt.Errorf("--azure-storage-account-key is required")
			}

			if managementConfig.PrintConfig {
				fmt.Println(tools.PrettyPrintStruct(managementConfig))
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			log.Infoln("Starting management")

			management.Run(&managementConfig)
		},
	}

	flags := cmd.PersistentFlags()

	flags.StringVarP(&cfgFile, "config", "c", "../../configs/management.yaml", "Config file path")

	flags.Int("management-port", 9000, "The management API port")
	viper.BindPFlag("management-port", flags.Lookup("management-port"))

	flags.String("namespace", "ion", "Namespace to deploy Ion into")
	cmd.MarkFlagRequired("namespace")
	viper.BindPFlag("namespace", flags.Lookup("namespace"))

	flags.String("dispatcher-image-name", "", "The container image name for the dispatcher")
	cmd.MarkFlagRequired("dispatcher-image-name")
	viper.BindPFlag("dispatcher-image-name", flags.Lookup("dispatcher-image-name"))

	flags.String("dispatcher-image-tag", "", "The container image tag")
	cmd.MarkFlagRequired("dispatcher-image-tag")
	viper.BindPFlag("dispatcher-image-tag", flags.Lookup("dispatcher-image-tag"))

	flags.String("azure-client-id", "", "Azure Service Principal Client ID")
	cmd.MarkFlagRequired("azure-client-id")
	viper.BindPFlag("azure-client-id", flags.Lookup("azure-client-id"))

	flags.String("azure-client-secret", "", "Azure Service Principal Client Secret")
	cmd.MarkFlagRequired("azure-client-secret")
	viper.BindPFlag("azure-client-secret", flags.Lookup("azure-client-secret"))

	flags.String("azure-subscription-id", "", "Azure Subscription ID")
	cmd.MarkFlagRequired("azure-subscription-id")
	viper.BindPFlag("azure-subscription-id", flags.Lookup("azure-subscription-id"))

	flags.String("azure-tenant-id", "", "Azure Tenant ID")
	viper.BindPFlag("azure-tenant-id", flags.Lookup("azure-tenant-id"))

	flags.String("azure-resource-group", "", "Azure Resource Group")
	viper.BindPFlag("azure-resource-group", flags.Lookup("azure-resource-group"))

	flags.String("azure-batch-pool-id", "", "Azure Batch Pool ID")
	viper.BindPFlag("azure-batch-pool-id", flags.Lookup("azure-batch-pool-id"))

	flags.String("azure-batch-account-name", "", "Azure Batch Account Name")
	viper.BindPFlag("azure-batch-account-name", flags.Lookup("azure-batch-account-name"))

	flags.String("azure-batch-account-location", "", "Azure Batch Account Location")
	viper.BindPFlag("azure-batch-account-location", flags.Lookup("azure-batch-account-location"))

	flags.String("mongodb-name", "", "MongoDB Name")
	viper.BindPFlag("mongodb-name", flags.Lookup("mongodb-name"))

	flags.String("mongodb-collection", "", "MongoDB Database Collection")
	viper.BindPFlag("mongodb-collection", flags.Lookup("mongodb-collection"))

	flags.String("mongodb-username", "", "MongoDB server username")
	viper.BindPFlag("mongodb-username", flags.Lookup("mongodb-username"))

	flags.String("mongodb-password", "", "MongoDB server password")
	viper.BindPFlag("mongodb-password", flags.Lookup("mongodb-password"))

	flags.Int("mongodb-port", 27017, "MongoDB server port")
	viper.BindPFlag("mongodb-port", flags.Lookup("mongodb-port"))

	flags.String("azure-storage-account-name", "", "ServiceBus topic name")
	viper.BindPFlag("azure-storage-account-name", flags.Lookup("azure-storage-account-name"))

	flags.String("azure-storage-account-key", "", "ServiceBus access key")
	viper.BindPFlag("azure-storage-account-key", flags.Lookup("azure-storage-account-key"))

	return cmd
}
