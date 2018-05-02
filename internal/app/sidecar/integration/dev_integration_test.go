package integration

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/lawrencegripper/ion/internal/app/sidecar/app"
	"github.com/lawrencegripper/ion/internal/app/sidecar/blob/filesystem"
	"github.com/lawrencegripper/ion/internal/app/sidecar/events/mock"
	"github.com/lawrencegripper/ion/internal/app/sidecar/meta/inmemory"
	"github.com/lawrencegripper/ion/internal/app/sidecar/types"
	"github.com/lawrencegripper/ion/internal/pkg/common"
	"github.com/sirupsen/logrus"
)

var sharedDB *inmemory.InMemoryDB

func TestDevIntegration(t *testing.T) {

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
	eventID := "1111111"

	config := &app.Configuration{
		SharedSecret: "secret",
		BaseDir:      baseDir,
		Context: &common.Context{
			Name:          "testmodule",
			EventID:       eventID,
			CorrelationID: "fish",
			ParentEventID: "",
		},
		ServerPort:  8080,
		PrintConfig: false,
		LogLevel:    "Debug",
		Development: true,
	}

	module1, err := createDevModule(config)
	if err != nil {
		t.Error(err)
	}
	defer module1.Close() // This is to ensure cleanup
	defer func() {
		_ = os.RemoveAll(types.DevBaseDir) // Clean up development files
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

	// Check dev.ready exists in development dir
	readyPath := path.Join(types.DevBaseDir, eventID, "dev.ready")
	if _, err := os.Stat(readyPath); os.IsNotExist(err) {
		t.Errorf("dev.ready file should exist at path '%s'", readyPath)
	}

	// Done will commit the written files to external providers
	if err := executeRequest(client, config.SharedSecret, fmt.Sprintf("%v", config.ServerPort), "done"); err != nil {
		t.Errorf("error calling done '%+v'", err)
	}

	// Check dev.done exists in development dir
	donePath := path.Join(types.DevBaseDir, eventID, "dev.done")
	if _, err := os.Stat(donePath); os.IsNotExist(err) {
		t.Errorf("dev.done file should exist at path '%s'", donePath)
	}

	// Clear local module directories
	module1.Close()

	// Hydrate event
	b, err := ioutil.ReadFile(path.Join(types.DevBaseDir, config.Context.EventID, "events", "event0.json"))
	if err != nil {
		t.Errorf("error reading event from disk '%+v'", err)
	}
	var inEvent common.Event
	err = json.Unmarshal(b, &inEvent)
	if err != nil {
		t.Errorf("error unmarshalling event '%+v'", err)
	}

	config.Context.ParentEventID = config.Context.EventID
	config.Context.EventID = inEvent.Context.EventID
	module2, err := createDevModule(config)
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

func createDevModule(config *app.Configuration) (*app.App, error) {
	if sharedDB == nil {
		var err error
		sharedDB, err = inmemory.NewInMemoryDB()
		if err != nil {
			return nil, fmt.Errorf("Failed to establish metadata store with debug provider, error: %+v", err)
		}
	}
	fsBlob, err := filesystem.NewBlobStorage(&filesystem.Config{
		InputDir:  path.Join(types.DevBaseDir, config.Context.ParentEventID, "blobs"),
		OutputDir: path.Join(types.DevBaseDir, config.Context.EventID, "blobs"),
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to establish metadata store with debug provider, error: %+v", err)
	}
	fsEvents := mock.NewEventPublisher(path.Join(types.DevBaseDir, config.Context.EventID, "events"))

	logger := logrus.New()
	logger.Out = os.Stdout

	a := app.App{}
	a.Setup(
		config.SharedSecret,
		config.BaseDir,
		config.Context,
		eventTypes,
		sharedDB,
		fsEvents,
		fsBlob,
		logger,
		config.Development,
	)

	go a.Run(fmt.Sprintf(":%d", config.ServerPort))
	return &a, nil
}
