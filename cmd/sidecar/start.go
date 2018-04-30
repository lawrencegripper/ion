package main

import (
	"errors"
	"fmt"

	"github.com/lawrencegripper/ion/internal/app/sidecar"
	"github.com/lawrencegripper/ion/internal/app/sidecar/app"
	"github.com/lawrencegripper/ion/internal/pkg/tools"
	"github.com/spf13/cobra"
)

// NewStartCommand create the start command with its flags
func NewStartCommand() *cobra.Command {
	config := app.Configuration{}
	var cmd = &cobra.Command{
		Use:   "start",
		Short: "ion-sidecar to embed in the processing",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// check emptyness of required parameters
			arr := []string{"azure-name", "azure-key", "azure-container", "mongo-name", "mongo-db", "bus-namespace", "bus-topic", "bus-key", "bus-rule-name"}
			for _, v := range arr {
				if sidecarCmdConfig.GetString(v) == "" {
					return errors.New("The parameter \"" + v + "\" cannot be empty")
				}
			}

			sidecarConfig.BaseDir = sidecarCmdConfig.GetString("base-dir")
			sidecarConfig.Context.Name = sidecarCmdConfig.GetString("module-name")
			sidecarConfig.ServerPort = sidecarCmdConfig.GetInt("port")

			sidecarConfig.AzureBlobProvider.BlobAccountName = sidecarCmdConfig.GetString("azure-name")
			sidecarConfig.AzureBlobProvider.BlobAccountKey = sidecarCmdConfig.GetString("azure-key")
			sidecarConfig.AzureBlobProvider.ContainerName = sidecarCmdConfig.GetString("azure-container")

			sidecarConfig.MongoDBMetaProvider.Name = sidecarCmdConfig.GetString("mongo-name")
			sidecarConfig.MongoDBMetaProvider.Password = sidecarCmdConfig.GetString("mongo-password")
			sidecarConfig.MongoDBMetaProvider.Collection = sidecarCmdConfig.GetString("mongo-db")
			sidecarConfig.MongoDBMetaProvider.Port = sidecarCmdConfig.GetInt("mongo-port")

			sidecarConfig.ServiceBusEventProvider.Namespace = sidecarCmdConfig.GetString("bus-namespace")
			sidecarConfig.ServiceBusEventProvider.Topic = sidecarCmdConfig.GetString("bus-topic")
			sidecarConfig.ServiceBusEventProvider.Key = sidecarCmdConfig.GetString("bus-key")
			sidecarConfig.ServiceBusEventProvider.AuthorizationRuleName = sidecarCmdConfig.GetString("bus-rule-name")

			if sidecarConfig.PrintConfig {
				fmt.Println(tools.PrettyPrintStruct(sidecarConfig))
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) { sidecar.Run(config) },
	}

	flags := cmd.PersistentFlags()
	flags.StringP("config", "c", "configs/sidecar.yml", "Path to the configuration file")
	flags.StringP("base-dir", "b", "./", "This base directory to use to store local files")
	flags.StringP("module-name", "n", "", "Module name")
	flags.String("azureblobprovider.blobaccountname", "", "Azure Blob Storage account name")
	flags.String("azureblobprovider.blobaccountkey", "", "Azure Blob Storage account key")
	flags.String("azureblobprovider.containername", "", "Azure Blob Storage container name")
	flags.String("mongodbmetaprovider.name", "", "MongoDB database name")
	flags.String("mongodbmetaprovider.password", "", "MongoDB database password")
	flags.String("mongodbmetaprovider.collection", "", "MongoDB database collection to use")
	flags.Int("mongodbmetaprovider.port", 27017, "MongoDB server port")
	flags.String("servicebuseventprovider.namespace", "", "ServiceBus namespace")
	flags.String("servicebuseventprovider.topic", "", "ServiceBus topic name")
	flags.String("servicebuseventprovider.key", "", "ServiceBus access key")
	flags.String("servicebuseventprovider.authorizationrulename", "", "ServiceBus authorization rule name")
	flags.IntP("port", "p", 8080, "Port to listen")

	sidecarCmdConfig.BindPFlag("base-dir", flags.Lookup("base-dir"))
	sidecarCmdConfig.BindPFlag("module-name", flags.Lookup("module-name"))
	sidecarCmdConfig.BindPFlag("azure-container", flags.Lookup("azureblobprovider.containername"))
	sidecarCmdConfig.BindPFlag("azure-name", flags.Lookup("azureblobprovider.blobaccountname"))
	sidecarCmdConfig.BindPFlag("azure-key", flags.Lookup("azureblobprovider.blobaccountkey"))
	sidecarCmdConfig.BindPFlag("mongo-name", flags.Lookup("mongodbmetaprovider.name"))
	sidecarCmdConfig.BindPFlag("mongo-password", flags.Lookup("mongodbmetaprovider.password"))
	sidecarCmdConfig.BindPFlag("mongo-db", flags.Lookup("mongodbmetaprovider.collection"))
	sidecarCmdConfig.BindPFlag("mongo-port", flags.Lookup("mongodbmetaprovider.port"))
	sidecarCmdConfig.BindPFlag("bus-namespace", flags.Lookup("servicebuseventprovider.namespace"))
	sidecarCmdConfig.BindPFlag("bus-topic", flags.Lookup("servicebuseventprovider.topic"))
	sidecarCmdConfig.BindPFlag("bus-key", flags.Lookup("servicebuseventprovider.key"))
	sidecarCmdConfig.BindPFlag("bus-rule-name", flags.Lookup("servicebuseventprovider.authorizationrulename"))
	sidecarCmdConfig.BindPFlag("port", flags.Lookup("port"))

	return cmd
}
