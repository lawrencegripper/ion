package links

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/satori/go.uuid"

	"github.com/lawrencegripper/ion/sidecar/events/servicebus"
	"github.com/lawrencegripper/ion/sidecar/types"
)

//Publisher is the event publisher frontapi will use for pushing newly arrived links
var publisher types.EventPublisher

func init() {
	publisher = getEventProvider()
}

//Process will read the URL inside the body of the incoming request and publish it to the topic
func Process(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var linkReq request
	err := decoder.Decode(&linkReq)
	if err != nil {
		log.Error(err)
	}
	defer r.Body.Close()

	log.Infoln("Processing URL:", linkReq.URL)

	uuid := uuid.Must(uuid.NewV4())

	event := types.Event{
		PreviousStages: []string{},
		CorrelationID:  uuid.String(),
		ParentEventID:  "",
		Data:           map[string]string{"url": linkReq.URL},
		Type:           "frontapi.new_link",
	}

	log.Infoln("Publishing event", publisher)
	err = publisher.Publish(event)
	if err != nil {
		log.Errorln(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	log.Infoln("Event published")

	// Send back a ressource_id
	fmt.Fprintln(w, "UUID:", uuid)
}

//TODO move this in a more generic pkg, outside of the app specifics
func getEventProvider() types.EventPublisher {
	c := servicebus.Config{
		Namespace: viper.GetString("servicebus_namespace"),
		Topic:     viper.GetString("servicebus_topic"),
		Key:       viper.GetString("servicebus_saspolicy"),
		AuthorizationRuleName: viper.GetString("servicebus_accesskey"),
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
