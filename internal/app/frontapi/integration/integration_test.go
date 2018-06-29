package integration

import (
	"bytes"
	"encoding/json"
	"github.com/lawrencegripper/ion/internal/app/frontapi"
	"github.com/lawrencegripper/ion/internal/app/handler/dataplane/documentstorage/mongodb"
	"github.com/lawrencegripper/ion/internal/pkg/common"
	"github.com/lawrencegripper/ion/internal/pkg/types"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"testing"
)

func TestHttpServer(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	var config = &types.Configuration{
		ClientID:            os.Getenv("AZURE_CLIENT_ID"),
		ClientSecret:        os.Getenv("AZURE_CLIENT_SECRET"),
		ResourceGroup:       os.Getenv("AZURE_RESOURCE_GROUP"),
		SubscriptionID:      os.Getenv("AZURE_SUBSCRIPTION_ID"),
		TenantID:            os.Getenv("AZURE_TENANT_ID"),
		ServiceBusNamespace: os.Getenv("AZURE_SERVICEBUS_NAMESPACE"),
		ModuleName:          "frontapi",
		SubscribesToEvent:   "none",
		EventsPublished:     "frontapi.new_link",
		LogLevel:            "Debug",
		Job: &types.JobConfig{
			RetryCount: 5,
		},
		Handler: &types.HandlerConfig{
			MongoDBDocumentStorageProvider: &types.MongoDBConfig{
				Collection: "inttest",
				Name:       os.Getenv("MONGODB_NAME"),
				Password:   os.Getenv("MONGODB_PASSWORD"),
				Port:       10255,
			},
		},
	}

	go frontapi.Run(config, 9898)

	i := 0

	for i < 6 {
		time.Sleep(time.Second * 10)
		i++

		req := struct {
			URL string `json:"url,omitEmpty"`
		}{
			URL: "http://doesnt.matter.not.real",
		}

		encodedReqBody, err := json.Marshal(req)
		if err != nil {
			t.Log(err)
			continue
		}

		request, err := http.NewRequest(http.MethodPost, "http://localhost:9898", bytes.NewBuffer(encodedReqBody))
		if err != nil {
			t.Log(err)
			continue
		}

		client := http.Client{}
		res, err := client.Do(request)
		if err != nil {
			t.Log(err)
			continue
		}

		resBytes, _ := ioutil.ReadAll(res.Body)
		if res.StatusCode == 200 {
			event := &common.Event{}
			json.Unmarshal(resBytes, event)
			checkEventData(t, event, config)
			return
		}
	}
}

func checkEventData(t *testing.T, e *common.Event, config *types.Configuration) {

	mongoConfig := &mongodb.Config{
		Enabled:    true,
		Name:       config.Handler.MongoDBDocumentStorageProvider.Name,
		Collection: config.Handler.MongoDBDocumentStorageProvider.Collection,
		Password:   config.Handler.MongoDBDocumentStorageProvider.Password,
		Port:       config.Handler.MongoDBDocumentStorageProvider.Port,
	}

	docStore, err := mongodb.NewMongoDB(mongoConfig)

	if err != nil {
		t.Error("Couldn't connect to mongo")
		return
	}

	res, err := docStore.GetEventMetaByID(e.Context.EventID)
	if err != nil {
		t.Error("Failed getting metadata from docstore")
		t.Error(err)
		return
	}

	if len(res.Data) != 1 {
		t.Error("metadata from docstore contained wrong number of items")
		return
	}

	if res.Data[0].Key != "url" || res.Data[0].Value != "http://doesnt.matter.not.real" {
		t.Errorf("data in docstore not as expected: %+v", res.Data[0])
	}
}
