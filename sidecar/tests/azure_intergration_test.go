package integration_tests

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"

	"github.com/lawrencegripper/ion/sidecar/app"
	"github.com/lawrencegripper/ion/sidecar/blob/azurestorage"
	"github.com/lawrencegripper/ion/sidecar/events/servicebus"
	"github.com/lawrencegripper/ion/sidecar/meta/mongodb"
	"github.com/sirupsen/logrus"
)

func TestAzureIntegration(t *testing.T) {

	mongoDBPort, err := strconv.ParseInt(os.Getenv("MONGODB_PORT"), 10, 32)
	if err != nil {
		panic("env var 'MONGODB_PORT' not set!")
	}

	config := &app.Configuration{
		SharedSecret:  "secret",
		ModuleName:    "testmodule",
		EventID:       "1111111",
		ExecutionID:   "123124",
		ParentEventID: "1111111",
		ServerPort:    8080,
		AzureBlobProvider: &azurestorage.Config{
			BlobAccountName: os.Getenv("AZURE_STORAGE_ACCOUNT_NAME"),
			BlobAccountKey:  os.Getenv("AZURE_STORAGE_ACCOUNT_KEY"),
			ContainerName:   "frank",
		},
		MongoDBMetaProvider: &mongodb.Config{
			Name:       os.Getenv("MONGODB_NAME"),
			Password:   os.Getenv("MONGODB_PASSWORD"),
			Collection: os.Getenv("MONGODB_COLLECTION"),
			Port:       int(mongoDBPort),
		},
		ServiceBusEventProvider: &servicebus.Config{
			Namespace: os.Getenv("SERVICEBUS_NAMESPACE"),
			Topic:     os.Getenv("SERVICEBUS_TOPIC"),
			Key:       os.Getenv("SERVICEBUS_KEY"),
			AuthorizationRuleName: os.Getenv("SERVICEBUS_AUTHRULENAME"),
		},
		PrintConfig: false,
		LogLevel:    "Debug",
	}

	db, err := mongodb.NewMongoDB(config.MongoDBMetaProvider)
	if err != nil {
		t.Errorf("failed to connect to mongodb with error '%+v'", err)
	}
	blob, err := azurestorage.NewBlobStorage(config.AzureBlobProvider, strings.Join([]string{
		config.EventID,
		config.ParentEventID,
		config.ModuleName}, "-"))
	if err != nil {
		t.Errorf("failed to connect to azure storage with error '%+v'", err)
	}
	sb, err := servicebus.NewServiceBus(config.ServiceBusEventProvider)
	if err != nil {
		t.Errorf("failed to connect to service bus with error '%+v'", err)
	}

	logger := logrus.New()
	logger.Out = os.Stdout

	a := app.App{}
	a.Setup(
		config.SharedSecret,
		config.EventID,
		config.ExecutionID,
		config.ParentEventID,
		config.ModuleName,
		db,
		sb,
		blob,
		true,
		logger,
	)

	defer a.Close()
	go a.Run(fmt.Sprintf(":%d", config.ServerPort))

	// Test on ready
	outDir := "out"
	dataDir := path.Join(outDir, "data")

	blob1 := "img1.png"
	blob1FilePath := path.Join(dataDir, blob1)
	err = ioutil.WriteFile(blob1FilePath, []byte("image1"), 0777)
	if err != nil {
		t.Errorf("error writing file '%s', '%+v'", blob1FilePath, err)
	}

	blob2 := "img2.png"
	blob2FilePath := path.Join(dataDir, blob2)
	err = ioutil.WriteFile(blob2FilePath, []byte("image2"), 0777)
	if err != nil {
		t.Errorf("error writing file '%s', '%+v'", blob2FilePath, err)
	}

	outFiles, err := ioutil.ReadDir(dataDir)
	if err != nil {
		t.Errorf("error reading out dir '%+v'", err)
	}
	outLength := len(outFiles)

	metadataJSONBytes := []byte("[{\"key\": \"key2\",\"value\": \"value2\"}]")
	metaFilePath := path.Join(outDir, "meta.json")
	err = ioutil.WriteFile(metaFilePath, metadataJSONBytes, 0777)
	if err != nil {
		t.Errorf("error opening metadata file '%s', '%+v'", metaFilePath, err)
	}

	json := fmt.Sprintf("{\"eventType\":\"face_detected\", \"parentEventID\":\"%s\", \"executionID\":\"%s\"}", config.ParentEventID, config.ExecutionID)
	eventJSONBytes := []byte(json)
	eventDir := path.Join(outDir, "events")
	eventFilePath := path.Join(eventDir, "event1.json")
	err = ioutil.WriteFile(eventFilePath, eventJSONBytes, 0777)
	if err != nil {
		t.Errorf("error opening event file '%s', '%+v'", metaFilePath, err)
	}

	doneURL := "http://localhost:" + fmt.Sprintf("%v", config.ServerPort) + "/done"
	req, err := http.NewRequest(http.MethodGet, doneURL, nil)
	req.Header.Set("secret", config.SharedSecret)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		t.Errorf("error calling done '%+v'", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("error code returned from done '%+v'", res.StatusCode)
	}

	readyURL, err := url.Parse("http://localhost:" + fmt.Sprintf("%v", config.ServerPort) + "/ready")
	req.URL = readyURL
	res, err = client.Do(req)
	if err != nil {
		t.Errorf("error calling ready '%+v'", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("error code returned from ready '%+v'", res.StatusCode)
	}

	// Check in/data is the same as out/data
	inDir := path.Join("in", "data")
	inFiles, err := ioutil.ReadDir(inDir)
	if err != nil {
		t.Errorf("error reading in dir '%+v'", err)
	}
	inLength := len(inFiles)

	if (inLength != outLength) && outLength > 0 {
		t.Errorf("error, input files length should match output length")
	}

	//TODO: event stuff
}
