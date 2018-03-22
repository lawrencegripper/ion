package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/containous/flaeg"
	log "github.com/sirupsen/logrus"
)

//Configuration represents the input configuration schema
type Configuration struct {
	SharedSecret            string `description:"A shared secret to authenticate client requests with"`
	BlobStorageAccessKey    string `description:"A access token for an external blob storage provider"`
	BlobStorageAccountName  string `description:"Blob storage account name"`
	DBName                  string `description:"The name of database to store metadata"`
	DBPassword              string `description:"The password to access the metadata database"`
	DBCollection            string `description:"The document collection name on the metadata database"`
	DBPort                  int    `description:"The database port"`
	PublisherName           string `description:"The name or namespace for the publisher"`
	PublisherTopic          string `description:"The topic to publish events on"`
	PublisherAccessKey      string `description:"An access key for the publisher"`
	PublisherAccessRuleName string `description:"The rule name associated with the given access key"`
	LogFile                 string `description:"File to log output to"`
	LogLevel                string `description:"Logging level, possible values {debug, info, warn, error}"`
}

func main() {

	config := &Configuration{}

	rootCmd := &flaeg.Command{
		Name:                  "start",
		Description:           "Creates a sidecar with helper services",
		Config:                config,
		DefaultPointersConfig: config,
		Run: func() error {
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
				config.PublisherAccessRuleName == "" {
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

func runApp(config *Configuration) {
	mongoDB, err := NewMongoDB(config.DBName, config.DBPassword, config.DBCollection, config.DBPort)
	if err != nil {
		panic(fmt.Errorf("Failed to connect to mongodb with error: %+v", err))
	}

	serviceBus, err := NewServiceBus(config.PublisherName, config.PublisherTopic, config.PublisherAccessKey, config.PublisherAccessRuleName)
	if err != nil {
		panic(fmt.Errorf("Failed to connect to servicebus with error: %+v", err))
	}

	azureBlob, err := NewAzureBlobStorage(config.BlobStorageAccountName, config.BlobStorageAccessKey)
	if err != nil {
		panic(fmt.Errorf("Failed to connect to azure blob storage with error: %+v", err))
	}

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

	app := App{}
	app.Setup(
		config.SharedSecret,
		config.BlobStorageAccessKey,
		mongoDB,
		serviceBus,
		azureBlob,
		logger,
	)

	defer app.Close()
	app.Run(":8080")
}

func prettyPrintStruct(item interface{}) string {
	b, _ := json.MarshalIndent(item, "", " ")
	return string(b)
}

func cleanConfig(c Configuration) Configuration {
	c.SharedSecret = "**********"
	c.BlobStorageAccessKey = "**********"
	c.DBPassword = "**********"
	c.PublisherAccessKey = "**********"
	return c
}
