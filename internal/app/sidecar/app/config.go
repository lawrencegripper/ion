package app

import (
	"github.com/lawrencegripper/ion/internal/app/sidecar/blob/azurestorage"
	"github.com/lawrencegripper/ion/internal/app/sidecar/events/servicebus"
	"github.com/lawrencegripper/ion/internal/app/sidecar/meta/mongodb"
	"github.com/lawrencegripper/ion/internal/pkg/common"
)

// cSpell:ignore mongodb

//Configuration represents the input configuration schema
type Configuration struct {
	SharedSecret            string               `description:"A shared secret to authenticate client requests with"`
	BaseDir                 string               `description:"This base directory to use to store local files"`
	Context                 *common.Context      `description:"The module details"`
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

func NewConfiguration() Configuration {
	cfg := Configuration{}
	cfg.Context = &common.Context{}
	cfg.AzureBlobProvider = &azurestorage.Config{}
	cfg.MongoDBMetaProvider = &mongodb.Config{}
	cfg.ServiceBusEventProvider = &servicebus.Config{}
	return cfg
}
