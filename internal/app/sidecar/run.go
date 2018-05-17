package sidecar

import (
	"fmt"
	"path/filepath"

	"os"
	"path"
	"runtime"
	"strings"

	"github.com/lawrencegripper/ion/internal/app/sidecar/committer"
	"github.com/lawrencegripper/ion/internal/app/sidecar/constants"
	"github.com/lawrencegripper/ion/internal/app/sidecar/dataplane"
	"github.com/lawrencegripper/ion/internal/app/sidecar/dataplane/blob/azurestorage"
	"github.com/lawrencegripper/ion/internal/app/sidecar/dataplane/blob/filesystem"
	"github.com/lawrencegripper/ion/internal/app/sidecar/dataplane/events/mock"
	"github.com/lawrencegripper/ion/internal/app/sidecar/dataplane/events/servicebus"
	"github.com/lawrencegripper/ion/internal/app/sidecar/dataplane/metadata/inmemory"
	"github.com/lawrencegripper/ion/internal/app/sidecar/dataplane/metadata/mongodb"
	"github.com/lawrencegripper/ion/internal/app/sidecar/helpers"
	"github.com/lawrencegripper/ion/internal/app/sidecar/preparer"
	log "github.com/sirupsen/logrus"
)

// cSpell:ignore flaeg, logrus, mongodb

const (
	defaultPort = 8080

	// A blank base dir will result in /ion/ being used
	defaultWindowsBaseDir = ""
	defaultLinuxBaseDir   = ""
	defaultDarwinBaseDir  = ""

	// Providers
	metaProviderMongoDB      string = "mongodb"
	metaProviderInMemory     string = "inmemory"
	blobProviderAzureStorage string = "azureblob"
	blobProviderFileSystem   string = "filesystem"
	eventProviderServiceBus  string = "servicebus"
)

// Run the sidecar using config
func Run(config Configuration) error {
	if err := validateConfig(&config); err != nil {
		return err
	}

	metaProvider := getMetaProvider(&config)
	blobProvider := getBlobProvider(&config)
	eventProvider := getEventProvider(&config)

	dataPlane := &dataplane.DataPlane{
		BlobProvider:     blobProvider,
		MetadataProvider: metaProvider,
		EventPublisher:   eventProvider,
	}

	logger := log.New()
	logger.Out = os.Stdout

	if config.LogFile != "" {
		file, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_WRONLY, 0666)
		if err == nil {
			logger.Out = file
		} else {
			logger.Info("Failed to log to file, using default stderr")
		}
	}

	logger.Level = mapLogLevel(config.LogLevel)

	validEventTypes := strings.Split(config.ValidEventTypes, ",")

	baseDir := config.BaseDir
	if baseDir == "" {
		baseDir = getDefaultBaseDir()
	}

	action := strings.ToLower(config.Action)
	if config.Action == constants.Prepare {
		preparer := preparer.NewPreparer(baseDir, config.Development, logger)
		defer preparer.Close()
		if err := preparer.Prepare(config.Context, dataPlane); err != nil {
			panic(fmt.Sprintf("Error during prepration %+v", err))
		}
	} else if config.Action == constants.Commit {
		committer := committer.NewCommitter(baseDir, config.Development, logger)
		defer committer.Close()
		if err := committer.Commit(config.Context, dataPlane, validEventTypes); err != nil {
			panic(fmt.Sprintf("Error during commit %+v", err))
		}
	} else {
		panic(fmt.Sprintf("Unsupported action type %+v", action))
	}

	return nil
}

func mapLogLevel(logLevel string) log.Level {
	switch strings.ToLower(logLevel) {
	case "debug":
		return log.DebugLevel
	case "info":
		return log.InfoLevel
	case "warn":
		return log.WarnLevel
	case "error":
		return log.ErrorLevel
	default:
		return log.WarnLevel
	}
}

func getDefaultBaseDir() string {
	switch runtime.GOOS {
	case "windows":
		return defaultWindowsBaseDir
	case "linux":
		return defaultLinuxBaseDir
	case "darwin":
		return defaultDarwinBaseDir
	default:
		panic("Unsupported OS platform")
	}
}

func validateConfig(c *Configuration) error {
	if (strings.ToLower(c.Action) != constants.Prepare &&
		strings.ToLower(c.Action) != constants.Commit) ||
		c.Context.EventID == "" ||
		c.Context.CorrelationID == "" {
		return fmt.Errorf("Missing or invalid configuration. Use '--printconfig' to show current config on start")
	}
	return nil
}

func getMetaProvider(config *Configuration) dataplane.MetadataProvider {
	if config.Development || config.MongoDBMetaProvider == nil {
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
			panic(fmt.Errorf("Failed to establish metadata store with provider '%s', error: %+v", metaProviderMongoDB, err))
		}
		return mongoDB
	}
	return nil
}

func getBlobProvider(config *Configuration) dataplane.BlobProvider {
	if config.Development || config.AzureBlobProvider == nil {
		fsBlob, err := filesystem.NewBlobStorage(&filesystem.Config{
			InputDir:  filepath.FromSlash(path.Join(constants.DevBaseDir, config.Context.ParentEventID, "blobs")),
			OutputDir: filepath.FromSlash(path.Join(constants.DevBaseDir, config.Context.EventID, "blobs")),
		})
		if err != nil {
			panic(fmt.Errorf("Failed to establish metadata store with debug provider, error: %+v", err))
		}
		return fsBlob
	}
	if config.AzureBlobProvider != nil {
		c := config.AzureBlobProvider
		azureBlob, err := azurestorage.NewBlobStorage(c,
			helpers.JoinBlobPath(config.Context.ParentEventID, config.Context.Name),
			helpers.JoinBlobPath(config.Context.EventID, config.Context.Name))
		if err != nil {
			panic(fmt.Errorf("Failed to establish blob storage with provider '%s', error: %+v", blobProviderAzureStorage, err))
		}
		return azureBlob
	}
	return nil
}

func getEventProvider(config *Configuration) dataplane.EventPublisher {
	if config.Development || config.ServiceBusEventProvider == nil {
		fsEvents := mock.NewEventPublisher(filepath.FromSlash(path.Join(constants.DevBaseDir, "events")))
		return fsEvents
	}
	if config.ServiceBusEventProvider != nil {
		c := config.ServiceBusEventProvider
		serviceBus, err := servicebus.NewServiceBus(c)
		if err != nil {
			panic(fmt.Errorf("Failed to establish event publisher with provider '%s', error: %+v", eventProviderServiceBus, err))
		}
		return serviceBus
	}
	return nil
}
