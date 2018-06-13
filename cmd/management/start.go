package main

import (
	"fmt"

	"github.com/lawrencegripper/ion/internal/app/management"
	"github.com/lawrencegripper/ion/internal/pkg/tools"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// NewStartCommand create the start command with its flags
func NewStartCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "start",
		Short: "ion-management to embed in the processing",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			managementConfig.Port = managementCmdConfig.GetInt("management-port")
			managementConfig.Namespace = managementCmdConfig.GetString("namespace")
			managementConfig.DispatcherImage = managementCmdConfig.GetString("dispatcher-image-name")
			managementConfig.DispatcherImageTag = managementCmdConfig.GetString("dispatcher-image-tag")
			managementConfig.AzureClientID = managementCmdConfig.GetString("azure-client-id")
			managementConfig.AzureClientSecret = managementCmdConfig.GetString("azure-client-secret")
			managementConfig.AzureSubscriptionID = managementCmdConfig.GetString("azure-subscription-id")
			managementConfig.AzureTenantID = managementCmdConfig.GetString("azure-tenant-id")
			managementConfig.AzureResourceGroup = managementCmdConfig.GetString("azure-resource-group")
			managementConfig.AzureBatchPoolID = managementCmdConfig.GetString("azure-batch-pool-id")
			managementConfig.AzureBatchAccountName = managementCmdConfig.GetString("azure-batch-account-name")
			managementConfig.AzureBatchAccountLocation = managementCmdConfig.GetString("azure-batch-account-location")
			managementConfig.AzureADResource = managementCmdConfig.GetString("azure-ad-resource")
			managementConfig.MongoDBName = managementCmdConfig.GetString("mongodb-name")
			managementConfig.MongoDBPort = managementCmdConfig.GetInt("mongodb-port")
			managementConfig.MongoDBCollection = managementCmdConfig.GetString("mongodb-collection")
			managementConfig.MongoDBPassword = managementCmdConfig.GetString("mongodb-password")
			managementConfig.AzureStorageAccountName = managementCmdConfig.GetString("azure-storage-account-name")
			managementConfig.AzureStorageAccountKey = managementCmdConfig.GetString("azure-storage-account-key")
			managementConfig.LogLevel = managementCmdConfig.GetString("loglevel")

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
	flags.Int("management-port", 9000, "The management API port")
	managementCmdConfig.BindPFlag("management-port", flags.Lookup("management-port"))

	flags.String("namespace", "ion", "Namespace to deploy Ion into")
	cmd.MarkFlagRequired("namespace")
	managementCmdConfig.BindPFlag("namespace", flags.Lookup("namespace"))

	flags.String("dispatcher-image-name", "", "The container image name for the dispatcher")
	cmd.MarkFlagRequired("dispatcher-image-name")
	managementCmdConfig.BindPFlag("dispatcher-image-name", flags.Lookup("dispatcher-image-name"))

	flags.String("dispatcher-image-tag", "", "The container image tag")
	cmd.MarkFlagRequired("dispatcher-image-tag")
	managementCmdConfig.BindPFlag("dispatcher-image-tag", flags.Lookup("dispatcher-image-tag"))

	flags.String("azure-client-id", "", "Azure Service Principal Client ID")
	cmd.MarkFlagRequired("azure-client-id")
	managementCmdConfig.BindPFlag(":azure-client-id", flags.Lookup(":azure-client-id"))

	flags.String("azure-client-secret", "", "Azure Service Principal Client Secret")
	cmd.MarkFlagRequired("azure-client-secret")
	managementCmdConfig.BindPFlag("azure-client-secret", flags.Lookup("azure-client-secret"))

	flags.String("azure-subscription-id", "", "Azure Subscription ID")
	managementCmdConfig.BindPFlag("azure-subscription-id", flags.Lookup("azure-subscription-id"))

	flags.String("azure-tenant-id", "", "Azure Tenant ID")
	managementCmdConfig.BindPFlag("azure-tenant-id", flags.Lookup("azure-tenant-id"))

	flags.String("azure-resource-group", "", "Azure Resource Group")
	managementCmdConfig.BindPFlag("azure-resource-group", flags.Lookup("azure-resource-group"))

	flags.String("azure-batch-pool-id", "", "Azure Batch Pool ID")
	managementCmdConfig.BindPFlag("azure-batch-pool-id", flags.Lookup("azure-batch-pool-id"))

	flags.String("azure-batch-account-name", "", "Azure Batch Account Name")
	managementCmdConfig.BindPFlag("azure-batch-account-name", flags.Lookup("azure-batch-account-name"))

	flags.String("azure-batch-account-location", "", "Azure Batch Account Location")
	managementCmdConfig.BindPFlag("azure-batch-account-location", flags.Lookup("azure-batch-account-location"))

	flags.String("azure-ad-resource", "", "Azure Active Directory Resource")
	managementCmdConfig.BindPFlag("azure-ad-resource", flags.Lookup("azure-ad-resource"))

	flags.String("mongodb-name", "", "MongoDB Name")
	managementCmdConfig.BindPFlag("mongodb-name", flags.Lookup("mongodb-name"))

	flags.String("mongodb-collection", "", "MongoDB Database Collection")
	managementCmdConfig.BindPFlag("mongodb-collection", flags.Lookup("mongodb-collection"))

	flags.String("mongodb-username", "", "MongoDB server username")
	managementCmdConfig.BindPFlag("mongodb-username", flags.Lookup("mongodb-username"))

	flags.String("mongodb-password", "", "MongoDB server password")
	managementCmdConfig.BindPFlag("mongodb-password", flags.Lookup("mongodb-password"))

	flags.Int("mongodb-port", 27017, "MongoDB server port")
	managementCmdConfig.BindPFlag("mongodb-port", flags.Lookup("mongodb-port"))

	flags.String("azure-storage-account-name", "", "ServiceBus topic name")
	managementCmdConfig.BindPFlag("azure-storage-account-name", flags.Lookup("azure-storage-account-name"))

	flags.String("azure-storage-account-key", "", "ServiceBus access key")
	managementCmdConfig.BindPFlag("azure-storage-account-key", flags.Lookup("azure-storage-account-key"))

	return cmd
}
