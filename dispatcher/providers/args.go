package providers

import (
	"strconv"

	"github.com/Azure/azure-sdk-for-go/services/servicebus/mgmt/2017-04-01/servicebus"
	"github.com/lawrencegripper/mlops/dispatcher/messaging"
	"github.com/lawrencegripper/mlops/dispatcher/types"
)

// GetSharedSidecarArgs gets the shared arguments used by the sidecar container
func GetSharedSidecarArgs(c *types.Configuration, sbKeys servicebus.AccessKeys) []string {
	return []string{
		"--azureblobprovider=true",
		"--azureblobprovider.blobaccountname=" + c.Sidecar.AzureBlobProvider.BlobAccountName,
		"--azureblobprovider.blobaccountkey=" + c.Sidecar.AzureBlobProvider.BlobAccountKey,
		"--azureblobprovider.useproxy=" + strconv.FormatBool(c.Sidecar.AzureBlobProvider.UseProxy),
		"--mongodbmetaprovider=true",
		"--mongodbmetaprovider.name=" + c.Sidecar.MongoDBMetaProvider.Name,
		"--mongodbmetaprovider.password=" + c.Sidecar.MongoDBMetaProvider.Password,
		"--mongodbmetaprovider.collection=" + c.Sidecar.MongoDBMetaProvider.Collection,
		"--mongodbmetaprovider.port=" + strconv.Itoa(c.Sidecar.MongoDBMetaProvider.Port),
		"--servicebuseventprovider=true",
		"--servicebuseventprovider.Namespace=" + c.ModuleName,
		"--servicebuseventprovider.Topic=" + c.EventsPublished,
		"--servicebuseventprovider.key=" + *sbKeys.PrimaryKey,
		"--servicebuseventprovider.authorizationrulename=" + *sbKeys.KeyName,
		"--serverport=" + strconv.Itoa(c.Sidecar.ServerPort),
		"--loglevel=" + c.LogLevel,
		"--printconfig=" + strconv.FormatBool(c.Sidecar.PrintConfig),
	}
}

func getMessageSidecarArgs(m messaging.Message) ([]string, error) {
	eventData, err := m.EventData()
	if err != nil {
		return []string{}, err
	}
	return []string{
		"--sharedsecret=" + m.ID(), //Todo: Investigate generating a more random secret
		"--eventid=" + m.ID(),
		"--correlationid=" + eventData.CorrelationID,
		"--parenteventid=" + eventData.ParentEventID,
	}, nil
}
