package integration

import (
	"encoding/json"
	"fmt"
	"github.com/lawrencegripper/ion/internal/app/handler"
	"github.com/lawrencegripper/ion/internal/app/handler/constants"
	"github.com/lawrencegripper/ion/internal/app/handler/dataplane/blobstorage/azure"
	"github.com/lawrencegripper/ion/internal/app/handler/dataplane/documentstorage/mongodb"
	"github.com/lawrencegripper/ion/internal/app/handler/development"
	"github.com/lawrencegripper/ion/internal/app/handler/module"
	"github.com/lawrencegripper/ion/internal/pkg/common"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// cSpell:ignore logrus, mongodb

var eventTypes = []string{
	"face_detected",
}

func TestAzureIntegration(t *testing.T) {

	if testing.Short() {
		t.Skip("Skipping integration test in short mode...")
	}

	devBaseDir := ".dev"
	devConfig := &development.Configuration{
		BaseDir: devBaseDir,
		Enabled: true,
	}

	// Create a phony a context for our
	// first test module
	eventID := "1111111"
	baseDir := "ion"
	eventTypesStr := "face_detected"
	eventTypes := strings.Split(eventTypesStr, ",")
	context := &common.Context{
		Name:          "testmodule",
		EventID:       eventID,
		CorrelationID: "fish",
		ParentEventID: "",
	}

	inEventsDir := filepath.FromSlash(filepath.Join(devBaseDir, eventID, devConfig.EventsDir()))
	inEventFilePath := filepath.FromSlash(path.Join(inEventsDir, "event0.json"))

	environment := module.GetModuleEnvironment(baseDir)

	mongoDBPort := os.Getenv("MONGODB_PORT")
	if mongoDBPort == "" {
		t.Fatal("env var 'MONGODB_PORT' not set!")
	}

	port, err := strconv.ParseInt(mongoDBPort, 10, strconv.IntSize)
	if err != nil {
		t.Fatal("env var 'MONGODB_PORT' should be an integer!")
	}

	// Build a configuration based on our phony context
	// and Azure environment variables
	config := handler.NewConfiguration()
	config.Action = constants.Prepare
	config.BaseDir = baseDir
	config.Context = context
	config.AzureBlobStorageProvider = &azure.Config{
		Enabled:         true,
		BlobAccountName: os.Getenv("AZURE_STORAGE_ACCOUNT_NAME"),
		BlobAccountKey:  os.Getenv("AZURE_STORAGE_ACCOUNT_KEY"),
		ContainerName:   "frank",
	}
	config.MongoDBDocumentStorageProvider = &mongodb.Config{
		Enabled:    true,
		Name:       os.Getenv("MONGODB_NAME"),
		Password:   os.Getenv("MONGODB_PASSWORD"),
		Collection: os.Getenv("MONGODB_COLLECTION"),
		Port:       int(port),
	}
	config.ValidEventTypes = eventTypesStr
	config.PrintConfig = false
	config.LogLevel = "Debug"
	config.DevelopmentConfiguration = devConfig

	// Run handler with 'Prepare' action  #1
	// This will create the 'ion' module env
	handler.Run(config)
	defer func() {
		_ = os.RemoveAll(baseDir) // This cleans up the local events directory created by the mock event publisher
		_ = os.RemoveAll(devBaseDir)
	}()

	// Write an output image blob
	blob1 := "img1.png"
	blob1FilePath := path.Join(environment.OutputBlobDirPath, blob1)
	writeOutputBlob(blob1FilePath)

	// Write an output image blob
	blob2 := "subdir/img2.png"
	blob2FilePath := path.Join(environment.OutputBlobDirPath, blob2)
	writeOutputBlob(blob2FilePath)

	// Grab the length of the output directory
	var outFiles []string
	err = filepath.Walk(environment.OutputBlobDirPath, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			return nil
		}
		outFiles = append(outFiles, path)
		return err
	})
	if err != nil {
		t.Fatalf("error walking input dir '%+v'", err)
	}
	outLength := len(outFiles)

	// Write an output metadata file
	insight := []byte(`[{"key": "key2","value": "value2"}]`)
	writeOutputBytes(insight, environment.OutputMetaFilePath)

	// Write an output event file
	j := fmt.Sprintf(`[{"key":"eventType","value":"%s"},{"key":"files","value":"%s,%s"},{"key":"abc","value":"123"}]`, eventTypes[0], blob1, blob2)
	outEvent := []byte(j)
	writeOutputBytes(outEvent, filepath.FromSlash(path.Join(environment.OutputEventsDirPath, "event1.json")))

	config.Action = constants.Commit
	// Run handler with 'commit' action #1
	// This will commit our data to the Azure
	// dataplane
	handler.Run(config)

	// Grab event ID from module 1's output event
	b, err := ioutil.ReadFile(inEventFilePath)
	if err != nil {
		t.Fatalf("error reading event from disk '%+v'", err)
	}
	var inEvent common.Event
	err = json.Unmarshal(b, &inEvent)
	if err != nil {
		t.Fatalf("error unmarshalling event '%+v'", err)
	}

	// Create Module #2
	config.Context.ParentEventID = config.Context.EventID
	config.Context.EventID = inEvent.Context.EventID
	config.Action = constants.Prepare
	handler.Run(config)

	// Check blob input data matches the output from the first module
	var inFiles []string
	err = filepath.Walk(environment.InputBlobDirPath, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			return nil
		}
		inFiles = append(inFiles, path)
		return err
	})
	if err != nil {
		t.Fatalf("error walking input dir '%+v'", err)
	}
	inLength := len(inFiles)

	if (inLength != outLength) && outLength > 0 {
		t.Fatal("error, input files length should match output length")
	}

	// Check the input metadata is the same as that output from the first module
	inMetaData, err := ioutil.ReadFile(environment.InputMetaFilePath)
	if err != nil {
		t.Fatalf("error reading in meta file '%s': '%+v'", environment.InputMetaFilePath, err)
	}

	var kvps common.KeyValuePairs
	err = json.Unmarshal(inMetaData, &kvps)
	if err != nil {
		t.Fatalf("error decoding file '%s' content: '%+v'", environment.InputMetaFilePath, err)
	}

	if len(kvps) != (inLength + 1) {
		t.Fatalf("insights file should contain %d key value pair, but has %d", (inLength + 1), len(kvps))
	}

	// The first key, value pair should be as expected
	kvp := kvps[0]
	if kvp.Key != "abc" {
		t.Fatalf("expected key 'abc' in key value pairs: '%+v'", kvp)
	}
	if kvp.Value != "123" {
		t.Fatalf("expected key 'abc' to have value '123' in key value pairs: '%+v'", kvp)
	}
}
