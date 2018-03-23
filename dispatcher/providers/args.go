package providers

import (
	"github.com/azure/azure-sdk-for-go/services/servicebus/mgmt/2017-04-01/servicebus"
	"github.com/lawrencegripper/mlops/dispatcher/messaging"
	"github.com/lawrencegripper/mlops/dispatcher/types"
)

// GetSharedSidecarArgs gets the shared arguments used by the sidecar container
func GetSharedSidecarArgs(c *types.Configuration, sbKeys servicebus.AccessKeys) []string {
	return []string{
		"--blobstorageaccesskey=" + c.Storage.BlobStorageAccessKey,
		"--blobstorageaccountname=" + c.Storage.BlobStorageName,
		"--dbname=" + c.Storage.MongoDbHostName,
		"--dbpassword=" + c.Storage.MongoDbPassword,
		"--dbcollection=" + c.Storage.MongoDbCollection,
		"--dbport=" + c.Storage.MongoDbPort,
		"--publishername=" + c.ModuleName,
		"--publisheraccesskey=" + *sbKeys.PrimaryKey,
		"--publisheraccessrulename=" + *sbKeys.KeyName,
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
