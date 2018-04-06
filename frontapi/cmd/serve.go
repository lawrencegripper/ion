package cmd

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/lawrencegripper/ion/frontapi/internal/app/frontapi/links"
)

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.PersistentFlags().Int("port", 8080, "Listenning port")

	viper.BindPFlag("port", serveCmd.PersistentFlags().Lookup("port"))
	viper.BindEnv("servicebus_namespace")
	viper.BindEnv("servicebus_topic")
	viper.BindEnv("servicebus_saspolicy")
	viper.BindEnv("servicebus_accesskey")

	viper.SetEnvPrefix("frontapi")
	viper.AutomaticEnv()
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve the HTTP handlers of frontapi",
	Run: func(cmd *cobra.Command, args []string) {
		// Routers declarations
		r := mux.NewRouter()

		// Routes handlings
		r.HandleFunc("/", links.Process).Methods("POST")

		// Server configuration
		server := &http.Server{
			Addr:    ":" + strconv.Itoa(viper.GetInt("port")), //TODO take this value from configuration file or from Cobra
			Handler: http.TimeoutHandler(r, 20*time.Second, "503 Service Unavailable"),
		}

		go func() {
			log.Infof("Starting listening: 0.0.0.0:%d", viper.GetInt("port")) //TODO take this value from configuration file or from Cobra
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
	},
}
