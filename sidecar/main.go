package main

import (
	"encoding/json"
	"fmt"

	"os"
	"path"
	"runtime"
	"strings"

	"github.com/containous/flaeg"
	"github.com/lawrencegripper/ion/sidecar/app"
	"github.com/lawrencegripper/ion/sidecar/blob/azurestorage"
	"github.com/lawrencegripper/ion/sidecar/blob/filesystem"
	"github.com/lawrencegripper/ion/sidecar/events/mock"
	"github.com/lawrencegripper/ion/sidecar/events/servicebus"
	"github.com/lawrencegripper/ion/sidecar/meta/inmemory"
	"github.com/lawrencegripper/ion/sidecar/meta/mongodb"
	"github.com/lawrencegripper/ion/sidecar/types"
	log "github.com/sirupsen/logrus"
)

// cSpell:ignore flaeg, logrus, mongodb

const (
	defaultPort = 8080

	// blank base dir will result in /ion/ being used
	defaultWindowsBaseDir = ""
	defaultLinuxBaseDir   = ""
	defaultDarwinBaseDir  = ""
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
			if err := validateConfig(config); err != nil {
				return err
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

	baseDir := config.BaseDir
	if baseDir == "" {
		switch runtime.GOOS {
		case "windows":
			baseDir = defaultWindowsBaseDir
		case "linux":
			baseDir = defaultLinuxBaseDir
		case "darwin":
			baseDir = defaultDarwinBaseDir
		default:
			//noop
		}
	}

	app := app.App{}
	app.Setup(
		config.SharedSecret,
		baseDir,
		config.Context,
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

func validateConfig(c *app.Configuration) error {
	if c.SharedSecret == "" || c.Context.EventID == "" || c.Context.CorrelationID == "" {
		return fmt.Errorf("Missing configuration. Use '--printconfig' to show current config on start")
	}
	//TODO: When more providers are added,
	// we need to check the configuration to
	// ensure only 1 for each type is set.
	// Alternatively, we allow multiple
	// provider configs and just return the
	// first we check against.
}

func getMetaProvider(config *app.Configuration) types.MetadataProvider {
	if config.Development {
		inMemDB, err := inmemory.NewInMemoryDB()
		if err != nil {
			panic(fmt.Errorf("Failed to establish metadata store with debug provider, error: %+v", err))
		}
		return inMemDB
	}
	if config.MongoDBMetaProvider != nil {
		c := config.MongoDBMetaProvider
		mongoDB, err := mongodb.NewMongoDB(c)
		if err != nil {
			panic(fmt.Errorf("Failed to establish metadata store with provider '%s', error: %+v", types.MetaProviderMongoDB, err))
		}
		return mongoDB
	}
	return nil
}

func getBlobProvider(config *app.Configuration) types.BlobProvider {
	if config.Development {
		fsBlob, err := filesystem.NewBlobStorage(&filesystem.Config{
			BaseDir: path.Join(types.DevBaseDir, "blobs"),
		})
		if err != nil {
			panic(fmt.Errorf("Failed to establish metadata store with debug provider, error: %+v", err))
		}
		return fsBlob
	}
	if config.AzureBlobProvider != nil {
		c := config.AzureBlobProvider
		azureBlob, err := azurestorage.NewBlobStorage(c,
			types.JoinBlobPath(config.Context.ParentEventID, config.Context.Name),
			types.JoinBlobPath(config.Context.EventID, config.Context.Name))
		if err != nil {
			panic(fmt.Errorf("Failed to establish blob storage with provider '%s', error: %+v", types.BlobProviderAzureStorage, err))
		}
		return azureBlob
	}
	return nil
}

func getEventProvider(config *app.Configuration) types.EventPublisher {
	if config.Development {
		fsEvents := mock.NewEventPublisher(path.Join(types.DevBaseDir, "events"))
		return fsEvents
	}
	if config.ServiceBusEventProvider != nil {
		c := config.ServiceBusEventProvider
		serviceBus, err := servicebus.NewServiceBus(c)
		if err != nil {
			panic(fmt.Errorf("Failed to establish event publisher with provider '%s', error: %+v", types.EventProviderServiceBus, err))
		}
		return serviceBus
	}
	return nil
}
