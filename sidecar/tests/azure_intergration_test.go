package integration_tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"

	"github.com/lawrencegripper/ion/sidecar/app"
	"github.com/lawrencegripper/ion/sidecar/blob/azurestorage"
	"github.com/lawrencegripper/ion/sidecar/events/mock"
	"github.com/lawrencegripper/ion/sidecar/meta/mongodb"
	"github.com/lawrencegripper/ion/sidecar/types"
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
		CorrelationID: "fish",
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
	sb := mock.NewMockEventPublisher("mockevents")

	logger := logrus.New()
	logger.Out = os.Stdout

	eventTypes := []string{
		"face_detected",
	}

	a := app.App{}
	a.Setup(
		config.SharedSecret,
		config.EventID,
		config.CorrelationID,
		config.ModuleName,
		eventTypes,
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

	// Write an output image blob
	blob1 := "img1.png"
	blob1FilePath := path.Join(dataDir, blob1)
	err = ioutil.WriteFile(blob1FilePath, []byte("image1"), 0777)
	if err != nil {
		t.Errorf("error writing file '%s', '%+v'", blob1FilePath, err)
	}

	// Write an output image blob
	blob2 := "img2.png"
	blob2FilePath := path.Join(dataDir, blob2)
	err = ioutil.WriteFile(blob2FilePath, []byte("image2"), 0777)
	if err != nil {
		t.Errorf("error writing file '%s', '%+v'", blob2FilePath, err)
	}

	// Grab the length of the output directory
	outFiles, err := ioutil.ReadDir(dataDir)
	if err != nil {
		t.Errorf("error reading out dir '%+v'", err)
	}
	outLength := len(outFiles)

	// Write an output metadata file
	metadataJSONBytes := []byte("[{\"key\": \"key2\",\"value\": \"value2\"}]")
	metaFilePath := path.Join(outDir, "meta.json")
	err = ioutil.WriteFile(metaFilePath, metadataJSONBytes, 0777)
	if err != nil {
		t.Errorf("error opening metadata file '%s', '%+v'", metaFilePath, err)
	}

	// Write an output event file
	j := fmt.Sprintf("[{\"key\":\"eventType\",\"value\":\"%s\"},{\"key\":\"files\",\"value\":\"%s,%s\"}]", eventTypes[0], blob1, blob2)
	eventJSONBytes := []byte(j)
	eventDir := path.Join(outDir, "events")
	eventFilePath := path.Join(eventDir, "event1.json")
	err = ioutil.WriteFile(eventFilePath, eventJSONBytes, 0777)
	if err != nil {
		t.Errorf("error opening event file '%s', '%+v'", metaFilePath, err)
	}

	client := &http.Client{}
	// Ready will attempt to sync the execution environment for this module - this should be empty
	firstReadyReq, err := http.NewRequest(http.MethodGet, "http://localhost:"+fmt.Sprintf("%v", config.ServerPort)+"/ready", nil)
	firstReadyReq.Header.Set("secret", config.SharedSecret)
	firstReadyRes, err := client.Do(firstReadyReq)
	if err != nil {
		t.Errorf("error calling ready '%+v'", err)
	}
	if firstReadyRes.StatusCode != http.StatusOK {
		t.Errorf("error code returned from ready '%+v'", firstReadyRes.StatusCode)
	}

	// Done will commit the written files to external providers
	doneReq, err := http.NewRequest(http.MethodGet, "http://localhost:"+fmt.Sprintf("%v", config.ServerPort)+"/done", nil)
	doneReq.Header.Set("secret", config.SharedSecret)
	doneRes, err := client.Do(doneReq)
	if err != nil {
		t.Errorf("error calling done '%+v'", err)
	}
	if doneRes.StatusCode != http.StatusOK {
		t.Errorf("error code returned from done '%+v'", doneRes.StatusCode)
	}

	eventPath := "mockevents/event0.json"
	b, err := ioutil.ReadFile(eventPath)
	if err != nil {
		t.Errorf("error reading event from disk '%+v'", err)
	}
	var event types.Event
	err = json.Unmarshal(b, &event)
	if err != nil {
		t.Errorf("error unmarshalling event '%+v'", err)
	}

	a2 := app.App{}
	a2.Setup(
		config.SharedSecret,
		event.EventID,
		config.CorrelationID,
		config.ModuleName,
		eventTypes,
		db,
		sb,
		blob,
		true,
		logger,
	)

	defer a2.Close()
	server2Port := fmt.Sprintf(":%d", config.ServerPort+1)
	go a2.Run(server2Port)

	// Ready will attempt to sync the execution environment for this module - this download the files written by the previous done
	secReadyReq, err := http.NewRequest(http.MethodGet, "http://localhost"+server2Port+"/ready", nil)
	secReadyReq.Header.Set("secret", config.SharedSecret)
	secReadyRes, err := client.Do(secReadyReq)
	if err != nil {
		t.Errorf("error calling ready '%+v'", err)
	}
	if secReadyRes.StatusCode != http.StatusOK {
		t.Errorf("error code returned from ready '%+v'", secReadyRes.StatusCode)
	}

	// Check inputs match outputs
	inDir := path.Join("in", "data")
	inFiles, err := ioutil.ReadDir(inDir)
	if err != nil {
		t.Errorf("error reading in dir '%+v'", err)
	}
	inLength := len(inFiles)

	if (inLength != outLength) && outLength > 0 {
		t.Errorf("error, input files length should match output length")
	}
}
