package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/containous/flaeg"
	"github.com/lawrencegripper/mlops/dispatcher/types"
)

func main() {
	log.Println("hello")
	hostName, err := os.Hostname()
	if err != nil {
		fmt.Println("Unable to automatically set instanceid to hostname")
	}

	config := &types.Configuration{
		Hostname: hostName,
	}

	rootCmd := &flaeg.Command{
		Name:                  "start",
		Description:           `Creates the dispatcher to process events`,
		Config:                config,
		DefaultPointersConfig: config,
		Run: func() error {
			fmt.Printf("Running dispatcher")
			if config.PrintConfig {
				fmt.Println(prettyPrintStruct(config))
			}
			if config.ClientID == "" || config.ClientSecret == "" || config.TenantID == "" || config.SubscriptionID == "" {
				panic("Missing configuration. Use '--printconfig' arg to show current config on start")
			}
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
