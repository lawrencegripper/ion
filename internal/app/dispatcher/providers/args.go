package providers

import (
	"github.com/Azure/azure-sdk-for-go/services/servicebus/mgmt/2017-04-01/servicebus"
	"github.com/joho/godotenv"
	"github.com/lawrencegripper/ion/internal/app/dispatcher/messaging"
	"github.com/lawrencegripper/ion/internal/pkg/common"
	"github.com/lawrencegripper/ion/internal/pkg/types"
	log "github.com/sirupsen/logrus"

	"os"
	"strconv"
)

// GetSharedSidecarArgs gets the shared arguments used by the sidecar container
func GetSharedSidecarArgs(c *types.Configuration, sbKeys servicebus.AccessKeys) []string {
	return []string{
		"--context.name=" + c.ModuleName,
		"--azureblobprovider=true",
		"--azureblobprovider.blobaccountname=" + c.Sidecar.AzureBlobProvider.BlobAccountName,
		"--azureblobprovider.blobaccountkey=" + c.Sidecar.AzureBlobProvider.BlobAccountKey,
		"--mongodbmetaprovider=true",
		"--mongodbmetaprovider.collection=" + c.Sidecar.MongoDBMetaProvider.Collection,
		"--mongodbmetaprovider.name=" + c.Sidecar.MongoDBMetaProvider.Name,
		"--mongodbmetaprovider.password=" + c.Sidecar.MongoDBMetaProvider.Password,
		"--mongodbmetaprovider.port=" + strconv.Itoa(c.Sidecar.MongoDBMetaProvider.Port),
		"--servicebuseventprovider=true",
		"--servicebuseventprovider.Namespace=" + c.ServiceBusNamespace,
		"--servicebuseventprovider.Topic=" + c.SubscribesToEvent,
		"--servicebuseventprovider.key=" + *sbKeys.PrimaryKey,
		"--servicebuseventprovider.authorizationrulename=" + *sbKeys.KeyName,
		"--serverport=" + strconv.Itoa(c.Sidecar.ServerPort),
		"--loglevel=" + c.LogLevel,
		"--printconfig=" + strconv.FormatBool(c.Sidecar.PrintConfig),
		"--valideventtypes=" + c.EventsPublished,
	}
}

func getMessageSidecarArgs(m messaging.Message) ([]string, error) {
	eventData, err := m.EventData()
	if err != nil {
		return []string{}, err
	}
	context := eventData.Context
	if context == nil {
		context = &common.Context{} // Use type defaults if no context
	}
	log.WithField("correlationid", context.CorrelationID).Debug("generating sidecar args for message")
	return []string{
		"--azureblobprovider.containername=" + context.CorrelationID,
		"--sharedsecret=" + m.ID(), //Todo: Investigate generating a more random secret
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
