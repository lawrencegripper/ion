package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"

	"github.com/lawrencegripper/ion/frontapi/links"
	"github.com/lawrencegripper/ion/sidecar/events/servicebus"
	"github.com/lawrencegripper/ion/sidecar/types"
)

func init() {
	links.Publisher = getEventProvider()
}

func main() {
	// Routers declarations
	r := mux.NewRouter()

	// Routes handlings
	r.HandleFunc("/", links.Process).Methods("POST")

	// Server configuration
	server := &http.Server{
		Addr:    ":" + strconv.Itoa(8080), //TODO take this value from configuration file or from Cobra
		Handler: http.TimeoutHandler(r, 20*time.Second, "503 Service Unavailable"),
	}

	go func() {
		log.Infof("Starting listening: 0.0.0.0:%d", 8080) //TODO take this value from configuration file or from Cobra
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalln(err)
		}
	}()

	// Catch signals for gracefully shutdown
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, syscall.SIGTERM, syscall.SIGINT)

	// Wait signal
	<-stopChan

	// Shutdown gracefully, but wait no longer than 20 seconds
	log.Infoln("Stopping gracefully the application...")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	server.Shutdown(ctx)

	links.Publisher.Close()

	log.Infoln("Application stopped gracefully")
}

func getEventProvider() types.EventPublisher {
	c := servicebus.Config{
		Namespace: "xxx",
		Topic:     "xxx",
		Key:       "xxx",
		AuthorizationRuleName: "xxx",
	}

	eventProviders := make([]types.EventPublisher, 0)
	// if config.ServiceBusEventProvider != nil {
	// c := config.ServiceBusEventProvider
	serviceBus, err := servicebus.NewServiceBus(&c)
	if err != nil {
		panic(fmt.Errorf("Failed to establish event publisher with provider '%s', error: %+v", types.EventProviderServiceBus, err))
	}
	eventProviders = append(eventProviders, serviceBus)
	// }
	// Do this rather than return a subset (first) of the providers to encourage quick failure
	if len(eventProviders) > 1 {
		panic("Only 1 metadata provider can be supplied")
	}
	if len(eventProviders) == 0 {
		panic("No metadata provider supplied, please add one.")
	}
	return eventProviders[0]
}
