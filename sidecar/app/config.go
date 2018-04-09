package app

import (
	"github.com/lawrencegripper/ion/sidecar/blob/azurestorage"
	"github.com/lawrencegripper/ion/sidecar/events/servicebus"
	"github.com/lawrencegripper/ion/sidecar/meta/mongodb"
	"github.com/lawrencegripper/ion/sidecar/types"
)

//Configuration represents the input configuration schema
type Configuration struct {
	SharedSecret            string               `description:"A shared secret to authenticate client requests with"`
	BaseDir                 string               `description:"This base directory to use to store local files"`
	Context                 *types.Context       `description:"The module details"`
	ValidEventTypes         string               `description:"Valid event type names as a comma delimited list"`
	ServerPort              int                  `description:"The port for the web server to listen on"`
	AzureBlobProvider       *azurestorage.Config `description:"Azure Storage Blob provider" export:"true"`
	MongoDBMetaProvider     *mongodb.Config      `description:"MongoDB metastore provider" export:"true"`
	ServiceBusEventProvider *servicebus.Config   `description:"ServiceBus event publisher" export:"true"`
	PrintConfig             bool                 `description:"Set to print config on start" export:"true"`
	LogFile                 string               `description:"File to log output to"`
	LogLevel                string               `description:"Logging level, possible values {debug, info, warn, error}"`
}

// cSpell:ignore mongodb
