package integration

import (
	"bytes"
	"encoding/json"
	"github.com/lawrencegripper/ion/internal/app/frontapi"
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
	success := false

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
			t.Log(string(resBytes))
			success = true
			break
		}

	}

	if !success {
		t.FailNow()
	}

}
