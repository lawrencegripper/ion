package handler

import (
	"github.com/lawrencegripper/ion/internal/app/handler/dataplane/blobstorage/azure"
	"github.com/lawrencegripper/ion/internal/app/handler/dataplane/documentstorage/mongodb"
	"github.com/lawrencegripper/ion/internal/app/handler/dataplane/events/servicebus"
	"github.com/lawrencegripper/ion/internal/pkg/common"
)

// cSpell:ignore mongodb

// Configuration represents the input Configuration schema
type Configuration struct {
	Action                         string             `description:"The action for the handler to perform (prepare or commit)"`
	BaseDir                        string             `description:"This base directory to use to store local files"`
	Context                        *common.Context    `description:"The module details"`
	ValidEventTypes                string             `description:"Valid event type names as a comma delimited list"`
	AzureBlobStorageProvider       *azure.Config      `description:"Azure Storage Blob provider" export:"true"`
	MongoDBDocumentStorageProvider *mongodb.Config    `description:"MongoDB metastore provider" export:"true"`
	ServiceBusEventProvider        *servicebus.Config `description:"ServiceBus event publisher" export:"true"`
	PrintConfig                    bool               `description:"Set to print config on start" export:"true"`
	LogFile                        string             `description:"File to log output to"`
	LogLevel                       string             `description:"Logging level, possible values {debug, info, warn, error}"`
	Development                    bool               `description:"A flag to enable development features"`
}

// NewConfiguration create an empty config
func NewConfiguration() Configuration {
	cfg := Configuration{}
	cfg.Context = &common.Context{}
	cfg.AzureBlobStorageProvider = &azure.Config{}
	cfg.MongoDBDocumentStorageProvider = &mongodb.Config{}
	cfg.ServiceBusEventProvider = &servicebus.Config{}
	return cfg
}
