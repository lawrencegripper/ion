package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/containous/flaeg"
	"github.com/lawrencegripper/mlops/dispatcher/providers/kubernetes"
	"github.com/lawrencegripper/mlops/dispatcher/servicebus"
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
			if config.LogSensitiveConfig {
				fmt.Println(prettyPrintStruct(config))
			} else {
				fmt.Println(prettyPrintStruct(types.RedactConfigSecrets(config)))
			}
			if config.ClientID == "" || config.ClientSecret == "" || config.TenantID == "" || config.SubscriptionID == "" {
				panic("Missing configuration. Use '--printconfig' arg to show current config on start")
			}

			ctx := context.Background()

			listener := servicebus.NewListener(ctx, config)
			for {
				message, err := listener.AmqpReceiver.Receive(ctx)
				if err != nil {
					// Todo: Investigate the type of error here. If this could be triggered by a poisened message
					// app shouldn't panic.
					log.WithError(err).Panic("Error received dequeuing message")
				}

				kubernetes.Dispatch(message, config)
			}
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