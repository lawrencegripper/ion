package main

import (
	"fmt"

	"github.com/lawrencegripper/ion/internal/app/sidecar"
	"github.com/lawrencegripper/ion/internal/pkg/tools"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// NewStartCommand create the start command with its flags
func NewStartCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "start",
		Short: "ion-sidecar to embed in the processing",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			sidecarConfig.BaseDir = sidecarCmdConfig.GetString("base-dir")
			sidecarConfig.Action = sidecarCmdConfig.GetString("action")
			sidecarConfig.ValidEventTypes = sidecarCmdConfig.GetString("valideventtypes")

			sidecarConfig.AzureBlobStorageProvider.Enabled = sidecarCmdConfig.GetBool("azureblobprovider.enabled")
			if sidecarConfig.AzureBlobStorageProvider.Enabled {
				sidecarConfig.AzureBlobStorageProvider.BlobAccountName = sidecarCmdConfig.GetString("azureblobprovider.blobaccountname")
				sidecarConfig.AzureBlobStorageProvider.BlobAccountKey = sidecarCmdConfig.GetString("azureblobprovider.blobaccountkey")
				sidecarConfig.AzureBlobStorageProvider.ContainerName = sidecarCmdConfig.GetString("azureblobprovider.containername")
			}

			sidecarConfig.MongoDBDocumentStorageProvider.Enabled = sidecarCmdConfig.GetBool("mongodbdocprovider.enabled")
			if sidecarConfig.MongoDBDocumentStorageProvider.Enabled {
				sidecarConfig.MongoDBDocumentStorageProvider.Name = sidecarCmdConfig.GetString("mongodbdocprovider.name")
				sidecarConfig.MongoDBDocumentStorageProvider.Password = sidecarCmdConfig.GetString("mongodbdocprovider.password")
				sidecarConfig.MongoDBDocumentStorageProvider.Collection = sidecarCmdConfig.GetString("mongodbdocprovider.collection")
				sidecarConfig.MongoDBDocumentStorageProvider.Port = sidecarCmdConfig.GetInt("mongodbdocprovider.port")
			}

			sidecarConfig.ServiceBusEventProvider.Enabled = sidecarCmdConfig.GetBool("servicebuseventprovider.enabled")
			if sidecarConfig.ServiceBusEventProvider.Enabled {
				sidecarConfig.ServiceBusEventProvider.Namespace = sidecarCmdConfig.GetString("servicebuseventprovider.namespace")
				sidecarConfig.ServiceBusEventProvider.Topic = sidecarCmdConfig.GetString("servicebuseventprovider.topic")
				sidecarConfig.ServiceBusEventProvider.Key = sidecarCmdConfig.GetString("servicebuseventprovider.key")
				sidecarConfig.ServiceBusEventProvider.AuthorizationRuleName = sidecarCmdConfig.GetString("servicebuseventprovider.authorizationrulename")
			}

			sidecarConfig.Context.Name = sidecarCmdConfig.GetString("context.name")
			sidecarConfig.Context.EventID = sidecarCmdConfig.GetString("context.eventid")
			sidecarConfig.Context.CorrelationID = sidecarCmdConfig.GetString("context.correlationid")
			sidecarConfig.Context.ParentEventID = sidecarCmdConfig.GetString("context.parenteventid")

			if sidecarConfig.PrintConfig {
				fmt.Println(tools.PrettyPrintStruct(sidecarConfig))
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			log.Infoln("Starting sidecar")

			sidecar.Run(sidecarConfig)
		},
	}

	flags := cmd.PersistentFlags()
	flags.StringP("base-dir", "b", "./", "This base directory to use to store local files")
	cmd.MarkFlagRequired("base-dir")
	sidecarCmdConfig.BindPFlag("base-dir", flags.Lookup("base-dir"))

	flags.StringP("action", "a", "", "The action for the sidecar to perform (prepare or commit)")
	cmd.MarkFlagRequired("action")
	sidecarCmdConfig.BindPFlag("action", flags.Lookup("action"))

	flags.String("valideventtypes", "", "Events which the module may raise on completion")
	cmd.MarkFlagRequired("valideventtypes")
	sidecarCmdConfig.BindPFlag("valideventtypes", flags.Lookup("valideventtypes"))

	flags.String("context.name", "", "Module name")
	cmd.MarkFlagRequired("context.name")
	sidecarCmdConfig.BindPFlag("context.name", flags.Lookup("context.name"))

	flags.String("context.eventid", "", "Event ID")
	cmd.MarkFlagRequired("context.eventid")
	sidecarCmdConfig.BindPFlag("context.eventid", flags.Lookup("context.eventid"))

	flags.String("context.correlationid", "", "Correlation ID")
	cmd.MarkFlagRequired("context.correlationid")
	sidecarCmdConfig.BindPFlag("context.correlationid", flags.Lookup("context.correlationid"))

	flags.String("context.parenteventid", "", "ParentEvent ID")
	sidecarCmdConfig.BindPFlag("context.parenteventid", flags.Lookup("context.parenteventid"))

	flags.Bool("azureblobprovider.enabled", false, "Enable Azure Blob Storage provider")
	sidecarCmdConfig.BindPFlag("azureblobprovider.enabled", flags.Lookup("azureblobprovider.enabled"))

	flags.String("azureblobprovider.blobaccountname", "", "Azure Blob Storage account name")
	sidecarCmdConfig.BindPFlag("azureblobprovider.blobaccountname", flags.Lookup("azureblobprovider.blobaccountname"))

	flags.String("azureblobprovider.blobaccountkey", "", "Azure Blob Storage account key")
	sidecarCmdConfig.BindPFlag("azureblobprovider.blobaccountkey", flags.Lookup("azureblobprovider.blobaccountkey"))

	flags.String("azureblobprovider.containername", "", "Azure Blob Storage container name")
	sidecarCmdConfig.BindPFlag("azureblobprovider.containername", flags.Lookup("azureblobprovider.containername"))

	flags.Bool("mongodbdocprovider.enabled", false, "Enable MongoDB Metadata provider")
	sidecarCmdConfig.BindPFlag("mongodbdocprovider.enabled", flags.Lookup("mongodbdocprovider.enabled"))

	flags.String("mongodbdocprovider.name", "", "MongoDB database name")
	sidecarCmdConfig.BindPFlag("mongodbdocprovider.name", flags.Lookup("mongodbdocprovider.name"))

	flags.String("mongodbdocprovider.password", "", "MongoDB database password")
	sidecarCmdConfig.BindPFlag("mongodbdocprovider.password", flags.Lookup("mongodbdocprovider.password"))

	flags.String("mongodbdocprovider.collection", "", "MongoDB database collection to use")
	sidecarCmdConfig.BindPFlag("mongodbdocprovider.collection", flags.Lookup("mongodbdocprovider.collection"))

	flags.Int("mongodbdocprovider.port", 27017, "MongoDB server port")
	sidecarCmdConfig.BindPFlag("mongodbdocprovider.port", flags.Lookup("mongodbdocprovider.port"))

	flags.Bool("servicebuseventprovider.enabled", false, "Enable Service Bus Event provider")
	sidecarCmdConfig.BindPFlag("servicebuseventprovider.enabled", flags.Lookup("servicebuseventprovider.enabled"))

	flags.String("servicebuseventprovider.namespace", "", "ServiceBus namespace")
	sidecarCmdConfig.BindPFlag("servicebuseventprovider.namespace", flags.Lookup("servicebuseventprovider.namespace"))

	flags.String("servicebuseventprovider.topic", "", "ServiceBus topic name")
	sidecarCmdConfig.BindPFlag("servicebuseventprovider.topic", flags.Lookup("servicebuseventprovider.topic"))

	flags.String("servicebuseventprovider.key", "", "ServiceBus access key")
	sidecarCmdConfig.BindPFlag("servicebuseventprovider.key", flags.Lookup("servicebuseventprovider.key"))

	flags.String("servicebuseventprovider.authorizationrulename", "", "ServiceBus authorization rule name")
	sidecarCmdConfig.BindPFlag("servicebuseventprovider.authorizationrulename", flags.Lookup("servicebuseventprovider.authorizationrulename"))

	return cmd
}
