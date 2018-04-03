package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/containous/flaeg"
	"github.com/lawrencegripper/ion/dispatcher/messaging"
	"github.com/lawrencegripper/ion/dispatcher/providers"
	"github.com/lawrencegripper/ion/dispatcher/servicebus"
	"github.com/lawrencegripper/ion/dispatcher/types"
)

func main() {
	log.Println("hello")
	hostName, err := os.Hostname()
	if err != nil {
		fmt.Println("Unable to automatically set instanceid to hostname")
	}

	config := &types.Configuration{
		Hostname: hostName,
		Job: &types.JobConfig{
			MaxRunningTimeMins: 10,
		},
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
				panic("Missing configuration. Use '--printconfig' or '-h' arg to show current config on start")
			}
			if config.Job == nil {
				panic("Job config can't be nil. Use '-h' to see options")
			}
			if config.Sidecar == nil {
				panic("Sidecar config can't be nil. Use '-h' to see options")
			}
			//Todo: validate sidecar config

			ctx := context.Background()

			listener := servicebus.NewListener(ctx, config)
			sidecarArgs := providers.GetSharedSidecarArgs(config, listener.AccessKeys)
			provider, err := providers.NewKubernetesProvider(config, sidecarArgs)
			if err != nil {
				log.WithError(err).Panic("Couldn't create kubernetes provider")
			}

			var wg sync.WaitGroup

			wg.Add(2)
			go func(wg *sync.WaitGroup) {
				defer wg.Done()
				for {
					message, err := listener.AmqpReceiver.Receive(ctx)
					if err != nil {
						// Todo: Investigate the type of error here. If this could be triggered by a poisened message
						// app shouldn't panic.
						log.WithError(err).Panic("Error received dequeuing message")
					}

					if message == nil {
						log.WithError(err).Panic("Error received dequeuing message - nil message")
					}

					err = provider.Dispatch(messaging.NewAmqpMessageWrapper(message))
					if err != nil {
						log.WithError(err).Error("Couldn't dispatch message to kubernetes provider")
					}
				}
			}(&wg)

			go func(wg *sync.WaitGroup) {
				defer wg.Done()
				for {
					err := provider.Reconcile()
					if err != nil {
						// Todo: Should this panic here? Should we tolerate a few failures (k8s upgade causing masters not to be vailable for example?)
						log.WithError(err).Panic("Failed to reconcile ....")
					}
					time.Sleep(time.Second * 15)
				}
			}(&wg)
			wg.Wait()

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
