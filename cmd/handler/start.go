// nolint: errcheck
package main

import (
	"fmt"

	"github.com/lawrencegripper/ion/internal/app/handler"
	"github.com/lawrencegripper/ion/internal/pkg/tools"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// NewStartCommand create the start command with its flags
func NewStartCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "start",
		Short: "ion-handler to embed in the processing",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			handlerConfig.BaseDir = handlerCmdConfig.GetString("base-dir")
			handlerConfig.Action = handlerCmdConfig.GetString("action")
			handlerConfig.ValidEventTypes = handlerCmdConfig.GetString("valideventtypes")

			handlerConfig.AzureBlobStorageProvider.Enabled = handlerCmdConfig.GetBool("azureblobprovider.enabled")
			if handlerConfig.AzureBlobStorageProvider.Enabled {
				handlerConfig.AzureBlobStorageProvider.BlobAccountName = handlerCmdConfig.GetString("azureblobprovider.blobaccountname")
				handlerConfig.AzureBlobStorageProvider.BlobAccountKey = handlerCmdConfig.GetString("azureblobprovider.blobaccountkey")
				handlerConfig.AzureBlobStorageProvider.ContainerName = handlerCmdConfig.GetString("azureblobprovider.containername")
			}

			handlerConfig.MongoDBDocumentStorageProvider.Enabled = handlerCmdConfig.GetBool("mongodbdocprovider.enabled")
			if handlerConfig.MongoDBDocumentStorageProvider.Enabled {
				handlerConfig.MongoDBDocumentStorageProvider.Name = handlerCmdConfig.GetString("mongodbdocprovider.name")
				handlerConfig.MongoDBDocumentStorageProvider.Password = handlerCmdConfig.GetString("mongodbdocprovider.password")
				handlerConfig.MongoDBDocumentStorageProvider.Collection = handlerCmdConfig.GetString("mongodbdocprovider.collection")
				handlerConfig.MongoDBDocumentStorageProvider.Port = handlerCmdConfig.GetInt("mongodbdocprovider.port")
			}

			handlerConfig.ServiceBusEventProvider.Enabled = handlerCmdConfig.GetBool("servicebuseventprovider.enabled")
			if handlerConfig.ServiceBusEventProvider.Enabled {
				handlerConfig.ServiceBusEventProvider.Namespace = handlerCmdConfig.GetString("servicebuseventprovider.namespace")
				handlerConfig.ServiceBusEventProvider.Topic = handlerCmdConfig.GetString("servicebuseventprovider.topic")
				handlerConfig.ServiceBusEventProvider.Key = handlerCmdConfig.GetString("servicebuseventprovider.key")
				handlerConfig.ServiceBusEventProvider.AuthorizationRuleName = handlerCmdConfig.GetString("servicebuseventprovider.authorizationrulename")
			}

			handlerConfig.Context.Name = handlerCmdConfig.GetString("context.name")
			handlerConfig.Context.EventID = handlerCmdConfig.GetString("context.eventid")
			handlerConfig.Context.CorrelationID = handlerCmdConfig.GetString("context.correlationid")
			handlerConfig.Context.ParentEventID = handlerCmdConfig.GetString("context.parenteventid")

			if handlerConfig.PrintConfig {
				fmt.Println(tools.PrettyPrintStruct(handlerConfig))
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			log.Infoln("Starting handler")

			handler.Run(handlerConfig)
		},
	}

	flags := cmd.PersistentFlags()
	flags.StringP("base-dir", "b", "./", "This base directory to use to store local files")
	cmd.MarkFlagRequired("base-dir")
	handlerCmdConfig.BindPFlag("base-dir", flags.Lookup("base-dir"))

	flags.StringP("action", "a", "", "The action for the handler to perform (prepare or commit)")
	cmd.MarkFlagRequired("action")
	handlerCmdConfig.BindPFlag("action", flags.Lookup("action"))

	flags.String("valideventtypes", "", "Events which the module may raise on completion")
	cmd.MarkFlagRequired("valideventtypes")
	handlerCmdConfig.BindPFlag("valideventtypes", flags.Lookup("valideventtypes"))

	flags.String("context.name", "", "Module name")
	cmd.MarkFlagRequired("context.name")
	handlerCmdConfig.BindPFlag("context.name", flags.Lookup("context.name"))

	flags.String("context.eventid", "", "Event ID")
	cmd.MarkFlagRequired("context.eventid")
	handlerCmdConfig.BindPFlag("context.eventid", flags.Lookup("context.eventid"))

	flags.String("context.correlationid", "", "Correlation ID")
	cmd.MarkFlagRequired("context.correlationid")
	handlerCmdConfig.BindPFlag("context.correlationid", flags.Lookup("context.correlationid"))

	flags.String("context.parenteventid", "", "ParentEvent ID")
	handlerCmdConfig.BindPFlag("context.parenteventid", flags.Lookup("context.parenteventid"))

	flags.Bool("azureblobprovider.enabled", false, "Enable Azure Blob Storage provider")
	handlerCmdConfig.BindPFlag("azureblobprovider.enabled", flags.Lookup("azureblobprovider.enabled"))

	flags.String("azureblobprovider.blobaccountname", "", "Azure Blob Storage account name")
	handlerCmdConfig.BindPFlag("azureblobprovider.blobaccountname", flags.Lookup("azureblobprovider.blobaccountname"))

	flags.String("azureblobprovider.blobaccountkey", "", "Azure Blob Storage account key")
	handlerCmdConfig.BindPFlag("azureblobprovider.blobaccountkey", flags.Lookup("azureblobprovider.blobaccountkey"))

	flags.String("azureblobprovider.containername", "", "Azure Blob Storage container name")
	handlerCmdConfig.BindPFlag("azureblobprovider.containername", flags.Lookup("azureblobprovider.containername"))

	flags.Bool("mongodbdocprovider.enabled", false, "Enable MongoDB Metadata provider")
	handlerCmdConfig.BindPFlag("mongodbdocprovider.enabled", flags.Lookup("mongodbdocprovider.enabled"))

	flags.String("mongodbdocprovider.name", "", "MongoDB database name")
	handlerCmdConfig.BindPFlag("mongodbdocprovider.name", flags.Lookup("mongodbdocprovider.name"))

	flags.String("mongodbdocprovider.password", "", "MongoDB database password")
	handlerCmdConfig.BindPFlag("mongodbdocprovider.password", flags.Lookup("mongodbdocprovider.password"))

	flags.String("mongodbdocprovider.collection", "", "MongoDB database collection to use")
	handlerCmdConfig.BindPFlag("mongodbdocprovider.collection", flags.Lookup("mongodbdocprovider.collection"))

	flags.Int("mongodbdocprovider.port", 27017, "MongoDB server port")
	handlerCmdConfig.BindPFlag("mongodbdocprovider.port", flags.Lookup("mongodbdocprovider.port"))

	flags.Bool("servicebuseventprovider.enabled", false, "Enable Service Bus Event provider")
	handlerCmdConfig.BindPFlag("servicebuseventprovider.enabled", flags.Lookup("servicebuseventprovider.enabled"))

	flags.String("servicebuseventprovider.namespace", "", "ServiceBus namespace")
	handlerCmdConfig.BindPFlag("servicebuseventprovider.namespace", flags.Lookup("servicebuseventprovider.namespace"))

	flags.String("servicebuseventprovider.topic", "", "ServiceBus topic name")
	handlerCmdConfig.BindPFlag("servicebuseventprovider.topic", flags.Lookup("servicebuseventprovider.topic"))

	flags.String("servicebuseventprovider.key", "", "ServiceBus access key")
	handlerCmdConfig.BindPFlag("servicebuseventprovider.key", flags.Lookup("servicebuseventprovider.key"))

	flags.String("servicebuseventprovider.authorizationrulename", "", "ServiceBus authorization rule name")
	handlerCmdConfig.BindPFlag("servicebuseventprovider.authorizationrulename", flags.Lookup("servicebuseventprovider.authorizationrulename"))

	return cmd
}
