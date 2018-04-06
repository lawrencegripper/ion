package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/containous/flaeg"
	"github.com/lawrencegripper/ion/sidecar/app"
	"github.com/lawrencegripper/ion/sidecar/blob/azurestorage"
	"github.com/lawrencegripper/ion/sidecar/events/servicebus"
	"github.com/lawrencegripper/ion/sidecar/meta/mongodb"
	"github.com/lawrencegripper/ion/sidecar/types"
	log "github.com/sirupsen/logrus"
)

const (
	defaultPort = 8080
)

func main() {

	config := &app.Configuration{}

	rootCmd := &flaeg.Command{
		Name:                  "start",
		Description:           "Creates a sidecar with helper services",
		Config:                config,
		DefaultPointersConfig: config,
		Run: func() error {
			addDefaults(config)
			fmt.Println("Running sidecar...")
			if config.PrintConfig {
				fmt.Println(prettyPrintStruct(*config))
			}
			if config.SharedSecret == "" || config.EventID == "" || config.CorrelationID == "" {
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

	validEventTypes := strings.Split(config.ValidEventTypes, ",")

	app := app.App{}
	app.Setup(
		config.SharedSecret,
		config.EventID,
		config.CorrelationID,
		config.ModuleName,
		validEventTypes,
		metaProvider,
		eventProvider,
		blobProvider,
		logger,
		config.Development,
	)

	defer app.Close()
	app.Run(fmt.Sprintf(":%d", config.ServerPort))
}

func prettyPrintStruct(item interface{}) string {
	b, _ := json.MarshalIndent(item, "", " ")
	return string(b)
}

func addDefaults(c *app.Configuration) {
	if c.ServerPort == 0 {
		c.ServerPort = defaultPort
	}
}

func getMetaProvider(config *app.Configuration) types.MetadataProvider {
	metaProviders := make([]types.MetadataProvider, 0)
	if config.MongoDBMetaProvider != nil {
		c := config.MongoDBMetaProvider
		mongoDB, err := mongodb.NewMongoDB(c)
		if err != nil {
			panic(fmt.Errorf("Failed to establish metadata store with provider '%s', error: %+v", types.MetaProviderMongoDB, err))
		}
		metaProviders = append(metaProviders, mongoDB)
	}
	// Do this rather than return a subset (first) of the providers to encourage quick failure
	if len(metaProviders) > 1 {
		panic("Only 1 metadata provider can be supplied")
	}
	if len(metaProviders) == 0 {
		panic("No metadata provider supplied, please add one.")
	}
	return metaProviders[0]
}

func getBlobProvider(config *app.Configuration) types.BlobProvider {
	blobProviders := make([]types.BlobProvider, 0)
	if config.AzureBlobProvider != nil {
		c := config.AzureBlobProvider
		azureBlob, err := azurestorage.NewBlobStorage(c, strings.Join([]string{
			config.EventID,
			config.ModuleName}, "-"))
		if err != nil {
			panic(fmt.Errorf("Failed to establish blob storage with provider '%s', error: %+v", types.BlobProviderAzureStorage, err))
		}
		blobProviders = append(blobProviders, azureBlob)
	}
	// Do this rather than return a subset (first) of the providers to encourage quick failure
	if len(blobProviders) > 1 {
		panic("Only 1 metadata provider can be supplied")
	}
	if len(blobProviders) == 0 {
		panic("No metadata provider supplied, please add one.")
	}
	return blobProviders[0]
}

func getEventProvider(config *app.Configuration) types.EventPublisher {
	eventProviders := make([]types.EventPublisher, 0)
	if config.ServiceBusEventProvider != nil {
		c := config.ServiceBusEventProvider
		serviceBus, err := servicebus.NewServiceBus(c)
		if err != nil {
			panic(fmt.Errorf("Failed to establish event publisher with provider '%s', error: %+v", types.EventProviderServiceBus, err))
		}
		eventProviders = append(eventProviders, serviceBus)
	}
	// Do this rather than return a subset (first) of the providers to encourage quick failure
	if len(eventProviders) > 1 {
		panic("Only 1 metadata provider can be supplied")
	}
	if len(eventProviders) == 0 {
		panic("No metadata provider supplied, please add one.")
	}
	return eventProviders[0]
}
