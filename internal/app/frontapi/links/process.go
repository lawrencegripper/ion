package links

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"

	"github.com/satori/go.uuid"

	//	"github.com/lawrencegripper/ion/internal/pkg/common"
)

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

	uuid := uuid.Must(uuid.NewV4(), nil)

	/*
		data := common.KeyValuePairs{}
		data.Append(common.KeyValuePair{Key: "url", Value: linkReq.URL})
			event := common.Event{
				PreviousStages: []string{},
				Data:           data,
				Type:           "frontapi.new_link",
			}

			context := common.Context{
				CorrelationID: uuid.String(),
				ParentEventID: "",
			}
				log.Infoln("Publishing event", publisher)
				err = publisher.Publish(event)
				if err != nil {
					log.Errorln(err)
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}
	*/
	log.Infoln("Event published")

	// Send back a ressource_id
	fmt.Fprintln(w, "UUID:", uuid)
}
