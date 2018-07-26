package providers

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/lawrencegripper/ion/internal/app/handler/dataplane/documentstorage"
	"github.com/lawrencegripper/ion/internal/app/handler/dataplane/documentstorage/mongodb"
	"github.com/lawrencegripper/ion/internal/pkg/messaging"
	"github.com/lawrencegripper/ion/internal/pkg/types"
	log "github.com/sirupsen/logrus"
	"time"
)

//LogStore captures modules logs
type LogStore struct {
	mongoStore   *mongodb.MongoDB
	blobStore    *storage.BlobStorageClient
	containerRef *storage.Container
	moduleName   string
}

//NewLogStore creates a new instance of the log store
func NewLogStore(mongoConfig *types.MongoDBConfig, blobConfig *types.AzureBlobConfig, moduleName string) (*LogStore, error) {
	logStore := LogStore{
		moduleName: moduleName,
	}
	if mongoConfig == nil || blobConfig == nil {
		return nil, fmt.Errorf("failed to create logstore, configuration missing")
	}

	mongoStore, err := mongodb.NewMongoDB(&mongodb.Config{
		Enabled:    true,
		Name:       mongoConfig.Name,
		Collection: mongoConfig.Collection,
		Password:   mongoConfig.Password,
		Port:       mongoConfig.Port,
	})
	if err != nil {
		return nil, fmt.Errorf("failed initialising mongo connection: %+v", err)
	}

	logStore.mongoStore = mongoStore

	blobClient, err := storage.NewBasicClient(blobConfig.BlobAccountName, blobConfig.BlobAccountKey)
	if err != nil {
		return nil, fmt.Errorf("failed initialising blob connection: %+v", err)
	}
	blobStore := blobClient.GetBlobService()
	logStore.blobStore = &blobStore

	logStore.containerRef = logStore.blobStore.GetContainerReference("logs")
	_, err = logStore.containerRef.CreateIfNotExists(&storage.CreateContainerOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create log container in blob: %+v", err)
	}

	return &logStore, nil
}

//StoreLogs persists logs to blob storage then creates a link to them in mongo
func (l *LogStore) StoreLogs(logger *log.Entry, message messaging.Message, stdout string, jobSuceeded bool) error {
	if l.mongoStore == nil || l.blobStore == nil {
		return errors.New("logstore not configured, failed to log messages")
	}

	eventData, err := message.EventData()
	if err != nil {
		logger.WithError(err).Error("failed to get eventData from message")
		return err
	}

	stdOutBuffer := bytes.NewBufferString(stdout)

	blobRef := l.containerRef.GetBlobReference(fmt.Sprintf("%s/%s/%s-attempt-%d.log", eventData.Context.CorrelationID, eventData.Context.EventID, eventData.Context.Name, message.DeliveryCount()))
	err = blobRef.CreateBlockBlobFromReader(stdOutBuffer, &storage.PutBlobOptions{})
	if err != nil {
		logger.WithError(err).Error("failed to get upload logs to blob")
		return err
	}

	readStorageOptions := storage.BlobSASOptions{
		BlobServiceSASPermissions: storage.BlobServiceSASPermissions{
			Read: true,
		},
		SASOptions: storage.SASOptions{
			Start:  time.Now().Add(time.Duration(-1) * time.Hour),
			Expiry: time.Now().Add(time.Duration(24) * time.Hour),
		},
	}
	sasURL, err := blobRef.GetSASURI(readStorageOptions)
	if err != nil {
		logger.WithError(err).Error("failed to store logs for job in blobstore")
		return err
	}

	// This event data will have the parent modules name, we want the logs to be stored under this module
	// so we update the context
	eventData.Context.Name = l.moduleName

	err = l.mongoStore.CreateModuleLogs(&documentstorage.ModuleLogs{
		Context:     eventData.Context,
		Logs:        sasURL,
		Succeeded:   jobSuceeded,
		Description: fmt.Sprintf("module:%s-event:%s-attempt:%v", eventData.Context.Name, eventData.Context.EventID, message.DeliveryCount()),
	})
	if err != nil {
		logger.WithError(err).Error("failed to store logs for job in metastore")
		return err
	}

	return nil
}
