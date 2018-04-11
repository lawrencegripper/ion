package integration

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"runtime"
	"strconv"
	"testing"

	"github.com/lawrencegripper/ion/common"
	"github.com/lawrencegripper/ion/sidecar/app"
	"github.com/lawrencegripper/ion/sidecar/blob/azurestorage"
	"github.com/lawrencegripper/ion/sidecar/events/mock"
	"github.com/lawrencegripper/ion/sidecar/meta/mongodb"
	"github.com/lawrencegripper/ion/sidecar/types"
	"github.com/sirupsen/logrus"
)

// cSpell:ignore logrus, mongodb

var eventTypes = []string{
	"face_detected",
}

func TestAzureIntegration(t *testing.T) {

	if testing.Short() {
		t.Skip("Skipping integration test in short mode...")
	}

	// Setting the base directory to empty
	// will result in /ion/... being used.
	baseDir := ""
	if runtime.GOOS == "windows" {
		// Use a relative base directory
		// on Windows to avoid Administrator
		// issues.
		baseDir = "ion"
	}
	outDir := path.Join(baseDir, "out")
	outDataDir := path.Join(outDir, "data")
	outMetaFilePath := path.Join(outDir, "meta.json")
	outEventsDir := path.Join(outDir, "events")
	outEventFilePath := path.Join(outEventsDir, "event1.json")
	inDir := path.Join(baseDir, "in")
	inDataDir := path.Join(inDir, "data")
	inMetaFilePath := path.Join(inDir, "meta.json")
	inEventsDir := "mockevents"
	inEventFilePath := path.Join(inEventsDir, "event0.json")

	mongoDBPort := os.Getenv("MONGODB_PORT")
	if mongoDBPort == "" {
		t.Errorf("env var 'MONGODB_PORT' not set!")
	}

	port, err := strconv.ParseInt(mongoDBPort, 10, strconv.IntSize)
	if err != nil {
		t.Errorf("env var 'MONGODB_PORT' should be an integer!")
	}

	config := &app.Configuration{
		SharedSecret: "secret",
		BaseDir:      baseDir,
		Context: &common.Context{
			Name:          "testmodule",
			EventID:       "1111111",
			CorrelationID: "fish",
			ParentEventID: "",
		},
		ServerPort: 8080,
		AzureBlobProvider: &azurestorage.Config{
			BlobAccountName: os.Getenv("AZURE_STORAGE_ACCOUNT_NAME"),
			BlobAccountKey:  os.Getenv("AZURE_STORAGE_ACCOUNT_KEY"),
			ContainerName:   "frank",
		},
		MongoDBMetaProvider: &mongodb.Config{
			Name:       os.Getenv("MONGODB_NAME"),
			Password:   os.Getenv("MONGODB_PASSWORD"),
			Collection: os.Getenv("MONGODB_COLLECTION"),
			Port:       int(port),
		},
		PrintConfig: false,
		LogLevel:    "Debug",
	}

	// Create Module #1
	module1, err := createModule(config)
	if err != nil {
		t.Error(err)
	}
	defer module1.Close() // This is to ensure cleanup
	defer func() {
		_ = os.RemoveAll(inEventsDir) // This cleans up the local events directory created by the mock event publisher
	}()

	// Write an output image blob
	blob1 := "img1.png"
	blob1FilePath := path.Join(outDataDir, blob1)
	writeOutputBlob(blob1FilePath)

	// Write an output image blob
	blob2 := "img2.png"
	blob2FilePath := path.Join(outDataDir, blob2)
	writeOutputBlob(blob2FilePath)

	// Grab the length of the output directory
	outFiles, err := ioutil.ReadDir(outDataDir)
	if err != nil {
		t.Errorf("error reading out dir '%+v'", err)
	}
	outLength := len(outFiles)

	// Write an output metadata file
	insight := []byte(`[{"key": "key2","value": "value2"}]`)
	writeOutputBytes(insight, outMetaFilePath)

	// Write an output event file
	j := fmt.Sprintf(`[{"key":"eventType","value":"%s"},{"key":"files","value":"%s,%s"},{"key":"abc","value":"123"}]`, eventTypes[0], blob1, blob2)
	outEvent := []byte(j)

	writeOutputBytes(outEvent, outEventFilePath)

	client := &http.Client{}

	// Ready will attempt to sync the execution environment for this module - this should be empty
	if err := executeRequest(client, config.SharedSecret, fmt.Sprintf("%v", config.ServerPort), "ready"); err != nil {
		t.Errorf("error calling ready '%+v'", err)
	}

	// Done will commit the written files to external providers
	if err := executeRequest(client, config.SharedSecret, fmt.Sprintf("%v", config.ServerPort), "done"); err != nil {
		t.Errorf("error calling done '%+v'", err)
	}

	// Clear local module directories
	module1.Close()

	// Hydrate event
	b, err := ioutil.ReadFile(inEventFilePath)
	if err != nil {
		t.Errorf("error reading event from disk '%+v'", err)
	}
	var inEvent common.Event
	err = json.Unmarshal(b, &inEvent)
	if err != nil {
		t.Errorf("error unmarshalling event '%+v'", err)
	}

	// Create Module #2
	config.Context.ParentEventID = config.Context.EventID
	config.Context.EventID = inEvent.Context.EventID
	module2, err := createModule(config)
	if err != nil {
		t.Error(err)
	}
	defer module2.Close()

	// Ready will attempt to sync the execution environment for this module.
	// This should download the files written by the previous done.
	if err := executeRequest(client, config.SharedSecret, fmt.Sprintf("%v", config.ServerPort), "ready"); err != nil {
		t.Errorf("error calling done '%+v'", err)
	}

	// Check blob input data matches the output from the first module
	inFiles, err := ioutil.ReadDir(inDataDir)
	if err != nil {
		t.Errorf("error reading in dir '%+v'", err)
	}
	inLength := len(inFiles)

	if (inLength != outLength) && outLength > 0 {
		t.Errorf("error, input files length should match output length")
	}

	// Check the input metadata is the same as that output from the first module
	inMetaData, err := ioutil.ReadFile(inMetaFilePath)
	if err != nil {
		t.Errorf("error reading in meta file '%s': '%+v'", inMetaFilePath, err)
	}

	var kvps common.KeyValuePairs
	err = json.Unmarshal(inMetaData, &kvps)
	if err != nil {
		t.Errorf("error decoding file '%s' content: '%+v'", inMetaFilePath, err)
	}

	// The first key, value pair should be as expected
	for _, kvp := range kvps {
		if kvp.Key != "abc" {
			t.Errorf("expected key 'abc' in key value pairs: '%+v'", kvp)
		}
		if kvp.Value != "123" {
			t.Errorf("expected key 'abc' to have value '123' in key value pairs: '%+v'", kvp)
		}
		break
	}
}

func createModule(config *app.Configuration) (*app.App, error) {
	db, err := mongodb.NewMongoDB(config.MongoDBMetaProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongodb with error '%+v'", err)
	}
	blob, err := azurestorage.NewBlobStorage(config.AzureBlobProvider,
		types.JoinBlobPath(config.Context.ParentEventID, config.Context.Name),
		types.JoinBlobPath(config.Context.EventID, config.Context.Name))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to azure storage with error '%+v'", err)
	}
	sb := mock.NewEventPublisher("mockevents")

	logger := logrus.New()
	logger.Out = os.Stdout

	a := app.App{}
	a.Setup(
		config.SharedSecret,
		config.BaseDir,
		config.Context,
		eventTypes,
		db,
		sb,
		blob,
		logger,
		true,
	)

	go a.Run(fmt.Sprintf(":%d", config.ServerPort))
	return &a, nil
}

func writeOutputBlob(path string) error {
	err := ioutil.WriteFile(path, []byte("image1"), 0777)
	if err != nil {
		return fmt.Errorf("error writing file '%s', '%+v'", path, err)
	}
	return nil
}

func writeOutputBytes(bytes []byte, path string) error {
	err := ioutil.WriteFile(path, bytes, 0777)
	if err != nil {
		return fmt.Errorf("error writing file '%s', '%+v'", bytes, err)
	}
	return nil
}

func executeRequest(client *http.Client, secret, port, path string) error {
	req, err := http.NewRequest(http.MethodGet, "http://localhost:"+port+"/"+path, nil)
	req.Header.Set("secret", secret)
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error calling '%s' '%+v'", path, err)
	}
	if res.StatusCode != http.StatusOK {
		var httpError types.ErrorResponse
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("unknown error returned from '%s'", path)
		}
		err = json.Unmarshal(b, &httpError)
		if err != nil {
			return fmt.Errorf("unknown error returned from '%s'", path)
		}
		return fmt.Errorf("error returned from '%s' '%+v'", path, httpError)
	}
	return nil
}
