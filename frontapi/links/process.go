package links

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/lawrencegripper/ion/sidecar/types"
	"github.com/satori/go.uuid"
)

var Publisher types.EventPublisher

func Process(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var linkReq Request
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

	err = Publisher.Publish(event)
	if err != nil {
		log.Errorln(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Send back a ressource_id
	fmt.Fprintln(w, "UUID:", uuid)
}
