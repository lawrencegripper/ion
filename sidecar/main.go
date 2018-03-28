package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/lawrencegripper/mlops/sidecar/meta/inmemory"

	"github.com/containous/flaeg"
	"github.com/lawrencegripper/mlops/sidecar/app"
	"github.com/lawrencegripper/mlops/sidecar/blob/azurestorage"
	"github.com/lawrencegripper/mlops/sidecar/blob/filesystem"
	"github.com/lawrencegripper/mlops/sidecar/events/servicebus"
	"github.com/lawrencegripper/mlops/sidecar/meta/mongodb"
	"github.com/lawrencegripper/mlops/sidecar/types"
	log "github.com/sirupsen/logrus"
	"github.com/vulcand/oxy/forward"
)

const hidden = "**********"
const defaultPort = 8080

//TODO: Currently this must respect the switch statements, when config is
//updated we can have a standard interface and configuration for providers
//which will release this coupling
const defaultBlobProvider = types.BlobProviderAzureStorage
const defaultMetaProvider = types.MetaProviderMongoDB
const defaultEventProvider = types.EventProviderServiceBus

func main() {

	config := &app.Configuration{}

	rootCmd := &flaeg.Command{
		Name:                  "start",
		Description:           "Creates a sidecar with helper services",
		Config:                config,
		DefaultPointersConfig: config,
		Run: func() error {
			addDefaults(config)
			fmt.Println("Running sidecar")
			fmt.Println("---------------")
			fmt.Println(prettyPrintStruct(cleanConfig(*config)))
			if config.SharedSecret == "" ||
				config.BlobStorageAccessKey == "" ||
				config.BlobStorageAccountName == "" ||
				config.DBName == "" ||
				config.DBPassword == "" ||
				config.DBCollection == "" ||
				config.DBPort == 0 ||
				config.PublisherName == "" ||
				config.PublisherTopic == "" ||
				config.PublisherAccessKey == "" ||
				config.PublisherAccessRuleName == "" ||
				config.EventID == "" ||
				config.ParentEventID == "" ||
				config.CorrelationID == "" {
				return fmt.Errorf("Missing configuration. Use '--printconfig' to show current config on start")
			}
			runApp(config)
			return nil
		},
	}

	flaeg := flaeg.New(rootCmd, os.Args[1:])

	if err := flaeg.Run(); err != nil {
		fmt.Printf("Error %s \n", err.Error())
	}
}

func runApp(config *app.Configuration) {

	//TODO: Sort configuration out - should be specific to providers
	metaProvider := getMetaProvider(config)
	blobProvider := getBlobProvider(config)
	eventProvider := getEventProvider(config)

	logger := log.New()
	logger.Out = os.Stdout

	if config.LogFile != "" {
		file, err := os.OpenFile("test.log", os.O_CREATE|os.O_WRONLY, 0666)
		if err == nil {
			logger.Out = file
		} else {
			logger.Info("Failed to log to file, using default stderr")
		}
	}

	switch strings.ToLower(config.LogLevel) {
	case "debug":
		logger.Level = log.DebugLevel
	case "info":
		logger.Level = log.InfoLevel
	case "warn":
		logger.Level = log.WarnLevel
	case "error":
		logger.Level = log.ErrorLevel
	default:
		logger.Level = log.WarnLevel
	}

	app := app.App{}
	app.Setup(
		config.SharedSecret,
		config.EventID,
		config.CorrelationID,
		config.ParentEventID,
		metaProvider,
		eventProvider,
		blobProvider,
		logger,
	)

	defer app.Close()
	app.Run(fmt.Sprintf(":%d", config.ServerPort))
}

func prettyPrintStruct(item interface{}) string {
	b, _ := json.MarshalIndent(item, "", " ")
	return string(b)
}

func cleanConfig(c app.Configuration) app.Configuration {
	c.SharedSecret = hidden
	c.BlobStorageAccessKey = hidden
	c.DBPassword = hidden
	c.PublisherAccessKey = hidden
	return c
}

func addDefaults(c *app.Configuration) {
	if c.ServerPort == 0 {
		c.ServerPort = defaultPort
	}
	if c.MetaProvider == "" {
		c.MetaProvider = defaultMetaProvider
	}
	if c.BlobProvider == "" {
		c.BlobProvider = defaultBlobProvider
	}
	if c.EventProvider == "" {
		c.EventProvider = defaultEventProvider
	}
}

func getMetaProvider(config *app.Configuration) types.MetaProvider {
	switch strings.ToLower(config.MetaProvider) {
	case types.MetaProviderMongoDB:
		mongoDB, err := mongodb.NewMongoDB(config.DBName, config.DBPassword, config.DBCollection, config.DBPort)
		if err != nil {
			panic(fmt.Errorf("Failed to establish metadata store with provider '%s', error: %+v", types.MetaProviderMongoDB, err))
		}
		return mongoDB
	case types.MetaProviderInMemory:
		inmemoryDB := inmemory.NewInMemoryMetaProvider(nil)
		return inmemoryDB
	default:
		mongoDB, err := mongodb.NewMongoDB(config.DBName, config.DBPassword, config.DBCollection, config.DBPort)
		if err != nil {
			panic(fmt.Errorf("Failed to connect to mongodb with error: %+v", err))
		}
		return mongoDB
	}
}

func getBlobProvider(config *app.Configuration) types.BlobProvider {
	switch strings.ToLower(config.BlobProvider) {
	case types.BlobProviderAzureStorage:
		//TODO: Proxy should be driven by config
		proxy, _ := forward.New(
			forward.Stream(true),
		)
		azureBlob, err := azurestorage.NewBlobStorage(config.BlobStorageAccountName, config.BlobStorageAccessKey, proxy)
		if err != nil {
			panic(fmt.Errorf("Failed to establish blob storage with provider '%s', error: %+v", types.BlobProviderAzureStorage, err))
		}
		return azureBlob
	case types.BlobProviderFileSystem:
		//TODO: baseDir driven by provider specific config
		filesystemBlob := filesystem.NewFileSystemBlobProvider("blobs")
		return filesystemBlob
	default:
		proxy, _ := forward.New(
			forward.Stream(true),
		)
		azureBlob, err := azurestorage.NewBlobStorage(config.BlobStorageAccountName, config.BlobStorageAccessKey, proxy)
		if err != nil {
			panic(fmt.Errorf("Failed to establish blob storage with provider '%s', error: %+v", types.BlobProviderAzureStorage, err))
		}
		return azureBlob
	}
}

func getEventProvider(config *app.Configuration) types.EventPublisher {
	switch strings.ToLower(config.EventProvider) {
	case types.EventProviderServiceBus:
		serviceBus, err := servicebus.NewServiceBus(config.PublisherName, config.PublisherTopic, config.PublisherAccessKey, config.PublisherAccessRuleName)
		if err != nil {
			panic(fmt.Errorf("Failed to establish event publisher with provider '%s', error: %+v", types.EventProviderServiceBus, err))
		}
		return serviceBus
	default:
		serviceBus, err := servicebus.NewServiceBus(config.PublisherName, config.PublisherTopic, config.PublisherAccessKey, config.PublisherAccessRuleName)
		if err != nil {
			panic(fmt.Errorf("Failed to establish event publisher with provider '%s', error: %+v", types.EventProviderServiceBus, err))
		}
		return serviceBus
	}
}
