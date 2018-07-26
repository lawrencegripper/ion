// nolint: errcheck
package main

import (
	"fmt"
	"strings"
	"time"

	logrus_appinsights "github.com/jjcollinge/logrus-appinsights"
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
				log.WithError(err).Warningln("Can't read management config from file %s", cfgFile)
			}
			viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
			viper.AutomaticEnv()

			managementConfig.Provider = viper.GetString("provider")
			managementConfig.Port = viper.GetInt("management-port")
			managementConfig.Namespace = viper.GetString("namespace")
			managementConfig.DispatcherImage = viper.GetString("dispatcher-image-name")
			managementConfig.AzureClientID = viper.GetString("azure-client-id")
			managementConfig.AzureClientSecret = viper.GetString("azure-client-secret")
			managementConfig.AzureSubscriptionID = viper.GetString("azure-subscription-id")
			managementConfig.AzureTenantID = viper.GetString("azure-tenant-id")
			managementConfig.AzureResourceGroup = viper.GetString("azure-resource-group")
			managementConfig.AzureBatchPoolID = viper.GetString("azure-batch-pool-id")
			managementConfig.AzureBatchAccountName = viper.GetString("azure-batch-account-name")
			managementConfig.AzureBatchAccountLocation = viper.GetString("azure-batch-account-location")
			managementConfig.AzureBatchRequiresGPU = viper.GetBool("azure-batch-requires-gpu")
			managementConfig.MongoDBName = viper.GetString("mongodb-name")
			managementConfig.MongoDBPort = viper.GetInt("mongodb-port")
			managementConfig.MongoDBCollection = viper.GetString("mongodb-collection")
			managementConfig.MongoDBPassword = viper.GetString("mongodb-password")
			managementConfig.AzureStorageAccountName = viper.GetString("azure-storage-account-name")
			managementConfig.AzureStorageAccountKey = viper.GetString("azure-storage-account-key")
			managementConfig.AzureServiceBusNamespace = viper.GetString("azure-servicebus-namespace")
			managementConfig.LogLevel = viper.GetString("loglevel")
			managementConfig.AppInsightsKey = viper.GetString("logging-appinsights")

			//Quick fix for https://github.com/lawrencegripper/ion/issues/142
			managementConfig.ContainerImageRegistryURL = viper.GetString("image-registry-url")
			managementConfig.ContainerImageRegistryUsername = viper.GetString("image-registry-username")
			managementConfig.ContainerImageRegistryPassword = viper.GetString("image-registry-password")

			managementConfig.AzureBatchImageRepositoryServer = viper.GetString("image-registry-url")
			managementConfig.AzureBatchImageRepositoryUsername = viper.GetString("image-registry-username")
			managementConfig.AzureBatchImageRepositoryPassword = viper.GetString("image-registry-password")

			if managementConfig.Provider == "" {
				return fmt.Errorf("--provider is required")
			}
			if managementConfig.DispatcherImage == "" {
				return fmt.Errorf("--dispatcher-image-name is required")
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
			// if managementConfig.AzureBatchPoolID == "" {
			// 	return fmt.Errorf("--azure-batch-pool-id is required")
			// }
			// if managementConfig.AzureBatchAccountName == "" {
			// 	return fmt.Errorf("--azure-batch-account-name is required")
			// }
			// if managementConfig.AzureBatchAccountLocation == "" {
			// 	return fmt.Errorf("--azure-batch-account-location is required")
			// }
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
			if managementConfig.AzureServiceBusNamespace == "" {
				return fmt.Errorf("--azure-servicebus-namespace is required")
			}

			if managementConfig.PrintConfig {
				fmt.Println(tools.PrettyPrintStruct(managementConfig))
			}

			if managementConfig.AppInsightsKey != "" {
				hook, err := logrus_appinsights.New("management-api", logrus_appinsights.Config{
					InstrumentationKey: managementConfig.AppInsightsKey,
					MaxBatchSize:       10,              // optional
					MaxBatchInterval:   time.Second * 5, // optional
				})
				if err != nil || hook == nil {
					panic(err)
				}

				// ignore fields
				hook.AddIgnore("private")
				log.AddHook(hook)
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

	flags.String("namespace", "default", "Namespace to deploy Ion into")
	cmd.MarkFlagRequired("namespace")
	viper.BindPFlag("namespace", flags.Lookup("namespace"))

	flags.String("dispatcher-image-name", "", "The container image name for the dispatcher")
	cmd.MarkFlagRequired("dispatcher-image-name")
	viper.BindPFlag("dispatcher-image-name", flags.Lookup("dispatcher-image-name"))

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

	flags.Bool("azure-batch-requires-gpu", true, "Azure Batch should use nvidia GPU")
	viper.BindPFlag("azure-batch-requires-gpu", flags.Lookup("azure-batch-requires-gpu"))

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

	flags.String("azure-storage-account-name", "", "Azure storage account name")
	viper.BindPFlag("azure-storage-account-name", flags.Lookup("azure-storage-account-name"))

	flags.String("azure-storage-account-key", "", "Azure storage account key")
	viper.BindPFlag("azure-storage-account-key", flags.Lookup("azure-storage-account-key"))

	flags.String("azure_servicebus_namespace", "", "Azure Service Bus namespace")
	viper.BindPFlag("azure_servicebus_namespace", flags.Lookup("azure_servicebus_namespace"))

	flags.String("image-registry-url", "", "The url of the image registry")
	viper.BindPFlag("image-registry-url", flags.Lookup("image-registry-url"))

	flags.String("image-registry-username", "", "The username for the image registry")
	viper.BindPFlag("image-registry-username", flags.Lookup("image-registry-url"))

	flags.String("image-registry-password", "", "The passworkd for the image registry")
	viper.BindPFlag("image-registry-password", flags.Lookup("image-registry-url"))

	//logging: Appinsights
	flags.String("logging-appinsights", "", "App Insights instrumentation key")
	viper.BindPFlag("logging-appinsights", flags.Lookup("logging-appinsights"))

	return cmd
}
