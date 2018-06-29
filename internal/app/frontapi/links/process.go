package links

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"pack.ag/amqp"

	log "github.com/sirupsen/logrus"

	"github.com/satori/go.uuid"

	"github.com/lawrencegripper/ion/internal/app/handler/dataplane/documentstorage"
	"github.com/lawrencegripper/ion/internal/pkg/common"
)

//Process will read the URL inside the body of the incoming request and publish it to the topic
func Process(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var linkReq request
	err := decoder.Decode(&linkReq)
	if err != nil {
		log.Error(err)
	}
	defer func() { _ = r.Body.Close() }()

	log.Infoln("Processing URL:", linkReq.URL)

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	event := common.Event{
		PreviousStages: []string{},
		Type:           eventType,
		Context: &common.Context{
			CorrelationID: uuid.Must(uuid.NewV4(), nil).String(),
			ParentEventID: "",
			EventID:       uuid.Must(uuid.NewV4(), nil).String(),
		},
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Errorf("failed marshalling event to json: %v", err)
		http.Error(w, "Failed marshalling event", http.StatusInternalServerError)
		return
	}

	// Create event metadata that
	// can store additional metadata
	// without bloating th event such
	// as a list of files to process.
	// This will be looked up by
	// the processing modules using the
	// event id.
	data := common.KeyValuePairs{}
	data = data.Append(common.KeyValuePair{Key: "url", Value: linkReq.URL})
	eventMeta := documentstorage.EventMeta{
		Context: event.Context,
		Data:    data,
	}
	err = documentStore.CreateEventMeta(&eventMeta)
	if err != nil {
		log.Errorf("failed to add context '%+v' with error '%+v'", eventMeta, err)
		http.Error(w, "Failed writing to document store", http.StatusInternalServerError)
		return
	}

	log.Infoln("Publishing event", amqpSender.Address())
	err = amqpSender.Send(ctx, &amqp.Message{
		Value: eventJSON,
	})
	if err != nil {
		log.Errorln(err)
		http.Error(w, "Failed publishing event", http.StatusInternalServerError)
		return
	}

	log.Infoln("Event published")

	// Send back a ressource_id
	err = json.NewEncoder(w).Encode(event)
	if err != nil {
		log.Errorln(err)
		http.Error(w, "Failed serialising result", http.StatusInternalServerError)
		return
	}
}
