package handler

import (
	"fmt"
	"github.com/lawrencegripper/ion/internal/app/handler/development"
	"path/filepath"

	"os"
	"path"
	"runtime"
	"strings"

	"github.com/lawrencegripper/ion/internal/app/handler/committer"
	"github.com/lawrencegripper/ion/internal/app/handler/constants"
	"github.com/lawrencegripper/ion/internal/app/handler/dataplane"
	"github.com/lawrencegripper/ion/internal/app/handler/dataplane/blobstorage/azure"
	"github.com/lawrencegripper/ion/internal/app/handler/dataplane/blobstorage/filesystem"
	"github.com/lawrencegripper/ion/internal/app/handler/dataplane/documentstorage/inmemory"
	"github.com/lawrencegripper/ion/internal/app/handler/dataplane/documentstorage/mongodb"
	"github.com/lawrencegripper/ion/internal/app/handler/dataplane/events/mock"
	"github.com/lawrencegripper/ion/internal/app/handler/dataplane/events/servicebus"
	"github.com/lawrencegripper/ion/internal/app/handler/helpers"
	"github.com/lawrencegripper/ion/internal/app/handler/preparer"
	log "github.com/sirupsen/logrus"
)

// cSpell:ignore flaeg, logrus, mongodb

const (
	// A blank base dir will result in /ion/ being used
	defaultWindowsBaseDir = ""
	defaultLinuxBaseDir   = ""
	defaultDarwinBaseDir  = ""

	// Providers
	metaProviderMongoDB      string = "mongodb"
	blobProviderAzureStorage string = "azureblob"
	eventProviderServiceBus  string = "servicebus"
)

// Run the handler using config
func Run(config Configuration) {
	if err := validateConfig(&config); err != nil {
		panic(err)
	}

	log.SetOutput(os.Stdout)
	if config.LogFile != "" {
		logFile, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_WRONLY, 0666)
		if err == nil {
			log.SetOutput(logFile)
		} else {
			log.Warnf("failed to open log file %s, using default stderr", config.LogFile)
		}
	}

	if config.DevelopmentConfiguration != nil {
		if config.DevelopmentConfiguration.Enabled {
			log.Debug("initializing development configuration")
			if err := config.DevelopmentConfiguration.Init(config.Context.ParentEventID, config.Context.EventID); err != nil {
				log.Errorf("%+v", err) // non fatal error
			}
		}
	}

	metaProvider := getMetaProvider(&config)
	blobProvider := getBlobProvider(&config)
	eventProvider := getEventProvider(&config)

	dataPlane := &dataplane.DataPlane{
		BlobStorageProvider:     blobProvider,
		DocumentStorageProvider: metaProvider,
		EventPublisher:          eventProvider,
	}

	// TODO Refactor out below into doRun(dataPlane *dataplane.Dataplane, config Configuration)

	validEventTypes := strings.Split(config.ValidEventTypes, ",")

	baseDir := config.BaseDir
	if baseDir == "" || baseDir == "./" || baseDir == ".\\" {
		baseDir = getDefaultBaseDir()
		log.Debugf("using default base directory %s", baseDir)
	}

	action := strings.ToLower(config.Action)
	if config.Action == constants.Prepare {
		preparer := preparer.NewPreparer(baseDir, config.DevelopmentConfiguration)
		defer preparer.Close()
		if err := preparer.Prepare(config.Context, dataPlane); err != nil {
			panic(fmt.Sprintf("error during prepration %+v", err))
		}
	} else if config.Action == constants.Commit {
		committer := committer.NewCommitter(baseDir, config.DevelopmentConfiguration)
		defer committer.Close()
		if err := committer.Commit(config.Context, dataPlane, validEventTypes); err != nil {
			panic(fmt.Sprintf("error during commit %+v", err))
		}
	} else {
		panic(fmt.Sprintf("unsupported action type %+v", action))
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
		panic("unsupported OS platform")
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

func getMetaProvider(config *Configuration) dataplane.DocumentStorageProvider {
	if config.MongoDBDocumentStorageProvider.Enabled {
		log.Info("using mongodb metadata provider")
		c := config.MongoDBDocumentStorageProvider
		mongoDB, err := mongodb.NewMongoDB(c)
		if err != nil {
			panic(fmt.Errorf("failed to establish metadata store with provider '%s', error: %+v", metaProviderMongoDB, err))
		}
		return mongoDB
	} // else
	log.Info("defaulting to in-memory metadata provider")
	inMemDB, err := inmemory.NewInMemoryDB()
	if err != nil {
		panic(fmt.Errorf("failed to establish metadata store with debug provider, error: %+v", err))
	}
	return inMemDB
}

func getBlobProvider(config *Configuration) dataplane.BlobStorageProvider {
	if config.AzureBlobStorageProvider.Enabled {
		log.Info("using azure blob storage provider")
		c := config.AzureBlobStorageProvider
		azureBlob, err := azure.NewBlobStorage(c,
			helpers.JoinBlobPath(config.Context.ParentEventID, config.Context.Name),
			helpers.JoinBlobPath(config.Context.EventID, config.Context.Name))
		if err != nil {
			panic(fmt.Errorf("failed to establish blob storage with provider '%s', error: %+v", blobProviderAzureStorage, err))
		}
		return azureBlob
	} // else
	log.Info("defaulting to filesystem blob storage provider")
	fsBlob, err := filesystem.NewBlobStorage(&filesystem.Config{
		InputDir:  filepath.FromSlash(path.Join(config.DevelopmentConfiguration.ParentModuleDir, development.BlobsDirExt)),
		OutputDir: filepath.FromSlash(path.Join(config.DevelopmentConfiguration.ModuleDir, development.BlobsDirExt)),
	})
	if err != nil {
		panic(fmt.Errorf("failed to establish metadata store with debug provider, error: %+v", err))
	}
	return fsBlob
}

func getEventProvider(config *Configuration) dataplane.EventPublisher {
	if config.ServiceBusEventProvider.Enabled {
		log.Info("using azure service bus event publisher")
		c := config.ServiceBusEventProvider
		serviceBus, err := servicebus.NewServiceBus(c)
		if err != nil {
			panic(fmt.Errorf("failed to establish event publisher with provider '%s', error: %+v", eventProviderServiceBus, err))
		}
		return serviceBus
	} // else
	log.Info("defaulting to filesystem event publisher")
	fsEvents := mock.NewEventPublisher(filepath.FromSlash(path.Join(config.DevelopmentConfiguration.ModuleDir, development.EventsDirExt)))
	return fsEvents
}
