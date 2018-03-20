package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/containous/flaeg"
)

// Configuration is a struct which contains all differents type to field
// using parsers on string, time.Duration, pointer, bool, int, int64, time.Time, float64
type Configuration struct {
	Name           string // no description struct tag, it will not be flaged
	LogLevel       string `short:"l" description:"Log level"` // string type field, short flag "-l"
	SubscriptionID string `description:"Timeout duration"`
	ClientID       string `description:"Timeout duration"`
	ClientSecret   string `description:"Timeout duration"`
	TenantID       string `description:"Timeout duration"`
}

func main() {
	log.Println("hello")
	hostName, err := os.Hostname()
	if err != nil {
		fmt.Println("Unable to automatically set instanceid to hostname")
	}

	config := &Configuration{
		Name: hostName,
	}

	rootCmd := &flaeg.Command{
		Name:                  "start",
		Description:           `Starts the watchdog, checking both the /health endpoint and request routing`,
		Config:                config,
		DefaultPointersConfig: config,
		Run: func() error {
			fmt.Printf("Running dispatcher")
			fmt.Println(prettyPrintStruct(config))
			return nil
		},
	}

	//init flaeg
	flaeg := flaeg.New(rootCmd, os.Args[1:])

	//run test
	if err := flaeg.Run(); err != nil {
		fmt.Printf("Error %s \n", err.Error())
	}
}

func prettyPrintStruct(item interface{}) string {
	b, _ := json.MarshalIndent(item, "", " ")
	return string(b)
}
