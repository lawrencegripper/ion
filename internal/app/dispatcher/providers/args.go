package providers

import (
	"github.com/Azure/azure-sdk-for-go/services/servicebus/mgmt/2017-04-01/servicebus"
	"github.com/joho/godotenv"
	"github.com/lawrencegripper/ion/internal/pkg/common"
	"github.com/lawrencegripper/ion/internal/pkg/messaging"
	"github.com/lawrencegripper/ion/internal/pkg/types"
	log "github.com/sirupsen/logrus"

	"os"
	"strconv"
)

// GetSharedHandlerArgs gets the shared arguments used by the handler container
func GetSharedHandlerArgs(c *types.Configuration, sbKeys servicebus.AccessKeys) []string {
	return []string{
		"--context.name=" + c.ModuleName,
		"--azureblobprovider.enabled=true",
		"--azureblobprovider.blobaccountname=" + c.Handler.AzureBlobStorageProvider.BlobAccountName,
		"--azureblobprovider.blobaccountkey=" + c.Handler.AzureBlobStorageProvider.BlobAccountKey,
		"--mongodbdocprovider.enabled=true",
		"--mongodbdocprovider.collection=" + c.Handler.MongoDBDocumentStorageProvider.Collection,
		"--mongodbdocprovider.name=" + c.Handler.MongoDBDocumentStorageProvider.Name,
		"--mongodbdocprovider.password=" + c.Handler.MongoDBDocumentStorageProvider.Password,
		"--mongodbdocprovider.port=" + strconv.Itoa(c.Handler.MongoDBDocumentStorageProvider.Port),
		"--servicebuseventprovider.enabled=true",
		"--servicebuseventprovider.namespace=" + c.ServiceBusNamespace,
		"--servicebuseventprovider.topic=" + c.SubscribesToEvent,
		"--servicebuseventprovider.key=" + *sbKeys.PrimaryKey,
		"--servicebuseventprovider.authorizationrulename=" + *sbKeys.KeyName,
		"--loglevel=" + c.LogLevel,
		"--printconfig=" + strconv.FormatBool(c.Handler.PrintConfig),
		"--valideventtypes=" + c.EventsPublished,
	}
}

func getMessageHandlerArgs(m messaging.Message) ([]string, error) {
	eventData, err := m.EventData()
	if err != nil {
		return []string{}, err
	}
	context := eventData.Context
	if context == nil {
		context = &common.Context{} // Use type defaults if no context
	}
	log.WithField("correlationid", context.CorrelationID).Debug("generating handler args for message")
	return []string{
		"--azureblobprovider.containername=" + context.CorrelationID,
		"--context.eventid=" + context.EventID,
		"--context.correlationid=" + context.CorrelationID,
		"--context.parenteventid=" + context.ParentEventID,
	}, nil
}

func getModuleEnvironmentVars(configLocation string) (map[string]string, error) {
	file, err := os.Open(configLocation)
	if err != nil {
		return map[string]string{}, err
	}
	// nolint:errcheck
	defer file.Close()
	envs, err := godotenv.Parse(file)
	return envs, err
}
