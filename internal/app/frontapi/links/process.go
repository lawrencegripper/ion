package links

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"pack.ag/amqp"

	log "github.com/sirupsen/logrus"

	"github.com/satori/go.uuid"

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

	uuid := uuid.Must(uuid.NewV4(), nil)

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	data := common.KeyValuePairs{}
	data.Append(common.KeyValuePair{Key: "url", Value: linkReq.URL})
	event := common.Event{
		Data: data,
		Type: "frontapi.new_link",
		Context: &common.Context{
			CorrelationID: uuid.String(),
			ParentEventID: "",
		},
	}

	log.Infoln("Publishing event", amqpClt.Sender.Address())
	err = amqpClt.Sender.Send(ctx, &amqp.Message{
		Value: event,
	})
	if err != nil {
		log.Errorln(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	log.Infoln("Event published")

	// Send back a ressource_id
	_, _ = fmt.Fprintln(w, "UUID:", uuid)
}
