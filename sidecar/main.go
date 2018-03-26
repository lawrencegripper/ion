package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/containous/flaeg"
	"github.com/lawrencegripper/mlops/sidecar/app"
	"github.com/lawrencegripper/mlops/sidecar/azure"
	"github.com/lawrencegripper/mlops/sidecar/common"
	log "github.com/sirupsen/logrus"
)

const hidden = "**********"

func main() {

	config := &common.Configuration{}

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

func runApp(config *common.Configuration) {
	mongoDB, err := azure.NewMongoDB(config.DBName, config.DBPassword, config.DBCollection, config.DBPort)
	if err != nil {
		panic(fmt.Errorf("Failed to connect to mongodb with error: %+v", err))
	}

	serviceBus, err := azure.NewServiceBus(config.PublisherName, config.PublisherTopic, config.PublisherAccessKey, config.PublisherAccessRuleName)
	if err != nil {
		panic(fmt.Errorf("Failed to connect to servicebus with error: %+v", err))
	}

	azureBlob, err := azure.NewAzureBlobStorage(config.BlobStorageAccountName, config.BlobStorageAccessKey)
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

	app := app.App{}
	app.Setup(
		config.SharedSecret,
		config.EventID,
		config.CorrelationID,
		config.ParentEventID,
		mongoDB,
		serviceBus,
		azureBlob,
		logger,
	)

	port := 8080
	if config.ServerPort != 0 {
		port = config.ServerPort
	}

	defer app.Close()
	app.Run(fmt.Sprintf(":%d", port))
}

func prettyPrintStruct(item interface{}) string {
	b, _ := json.MarshalIndent(item, "", " ")
	return string(b)
}

func cleanConfig(c common.Configuration) common.Configuration {
	c.SharedSecret = hidden
	c.BlobStorageAccessKey = hidden
	c.DBPassword = hidden
	c.PublisherAccessKey = hidden
	return c
}
