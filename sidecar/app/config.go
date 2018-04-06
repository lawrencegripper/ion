package app

import (
	"github.com/lawrencegripper/ion/sidecar/blob/azurestorage"
	"github.com/lawrencegripper/ion/sidecar/events/servicebus"
	"github.com/lawrencegripper/ion/sidecar/meta/mongodb"
)

//Configuration represents the input configuration schema
type Configuration struct {
	SharedSecret            string               `description:"A shared secret to authenticate client requests with"`
	ModuleName              string               `description:"The module's name"`
	EventID                 string               `description:"The unique ID for this module"`
	CorrelationID           string               `description:"The correlation ID"`
	ValidEventTypes         string               `description:"Valid event type names as a comma delimited list"`
	ServerPort              int                  `description:"The port for the web server to listen on"`
	AzureBlobProvider       *azurestorage.Config `description:"Azure Storage Blob provider" export:"true"`
	MongoDBMetaProvider     *mongodb.Config      `description:"MongoDB metastore provider" export:"true"`
	ServiceBusEventProvider *servicebus.Config   `description:"ServiceBus event publisher" export:"true"`
	PrintConfig             bool                 `description:"Set to print config on start" export:"true"`
	LogFile                 string               `description:"File to log output to"`
	LogLevel                string               `description:"Logging level, possible values {debug, info, warn, error}"`
	Development             bool                 `description:"A flag to enable development features"`
}
