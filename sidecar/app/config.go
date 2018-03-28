package app

import (
	"github.com/lawrencegripper/mlops/sidecar/blob/azurestorage"
	"github.com/lawrencegripper/mlops/sidecar/blob/filesystem"
	"github.com/lawrencegripper/mlops/sidecar/events/servicebus"
	"github.com/lawrencegripper/mlops/sidecar/meta/mongodb"
)

//Configuration represents the input configuration schema
type Configuration struct {
	SharedSecret            string               `description:"A shared secret to authenticate client requests with"`
	LogFile                 string               `description:"File to log output to"`
	LogLevel                string               `description:"Logging level, possible values {debug, info, warn, error}"`
	EventID                 string               `description:"The unique ID for this module"`
	ParentEventID           string               `description:"Previous event ID"`
	CorrelationID           string               `description:"CorrelationID used to correlate this module with others"`
	ServerPort              int                  `description:"The port for the web server to listen on"`
	FileSystemBlobProvider  *filesystem.Config   `description:"File system blob provider" export:"true"`
	AzureBlobProvider       *azurestorage.Config `description:"Azure Storage Blob provider" export:"true"`
	MongoDBMetaProvider     *mongodb.Config      `description:"MongoDB metastore provider" export:"true"`
	ServiceBusEventProvider *servicebus.Config   `description:"ServiceBus event publisher" export:"true"`
	PrintConfig             bool                 `description:"Set to print config on start" export:"true"`
}
