package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/containous/flaeg"
)

type Configuration struct {
	SharedSecret            string `description:"A shared secret to authenticate client requests with"`
	BlobStorageAccessKey    string `description:"A access token for an external blob storage provider"`
	DBName                  string `description:"The name of database to store metadata"`
	DBPassword              string `description:"The password to access the metadata database"`
	DBCollection            string `description:"The document collection name on the metadata database"`
	DBPort                  int    `description:"The database port"`
	PublisherName           string `description:"The name or namespace for the publisher"`
	PublisherTopic          string `description:"The topic to publish events on"`
	PublisherAccessKey      string `description:"An access key for the publisher"`
	PublisherAccessRuleName string `description:"The rule name associated with the given access key"`
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
			fmt.Println(prettyPrintStruct(config))
			if config.SharedSecret == "" ||
				config.BlobStorageAccessKey == "" ||
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
		panic(fmt.Errorf("Failed to create mongodb with error: %+v", err))
	}

	serviceBus, err := NewServiceBus(config.PublisherName, config.PublisherTopic, config.PublisherAccessKey, config.PublisherAccessRuleName)
	if err != nil {
		panic(fmt.Errorf("Failed to create servicebus with error: %+v", err))
	}

	a := App{}
	a.Setup(
		config.SharedSecret,
		config.BlobStorageAccessKey,
		mongoDB,
		serviceBus,
	)

	defer a.Close()
	a.Run(":8080")
}

func prettyPrintStruct(item interface{}) string {
	b, _ := json.MarshalIndent(item, "", " ")
	return string(b)
}
