package frontapi

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/lawrencegripper/ion/internal/app/frontapi/links"
	"github.com/lawrencegripper/ion/internal/pkg/types"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// Run starts the webserver that on port
func Run(cfg *types.Configuration, port int) {

	links.InitAmqp(cfg)

	// Routers declarations
	r := mux.NewRouter()

	// Routes handlings
	r.HandleFunc("/", links.Process).Methods("POST")

	// Server configuration
	server := &http.Server{
		Addr:    ":" + strconv.Itoa(port), //TODO take this value from configuration file or from Cobra
		Handler: http.TimeoutHandler(r, 20*time.Second, "503 Service Unavailable"),
	}

	go func() {
		log.Infof("Starting listening: 0.0.0.0:%d", port) //TODO take this value from configuration file or from Cobra
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

	//TODO When should I call that?
	//links.publisher.Close()

	log.Infoln("Application stopped gracefully")
}
