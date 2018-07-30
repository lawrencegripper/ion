package integration

import (
	"encoding/json"
	"fmt"
	"github.com/lawrencegripper/ion/internal/app/handler/development"
	"github.com/lawrencegripper/ion/internal/app/handler/module"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"testing"

	"github.com/lawrencegripper/ion/internal/app/handler"
	"github.com/lawrencegripper/ion/internal/app/handler/constants"
	"github.com/lawrencegripper/ion/internal/pkg/common"
)

func TestDevIntegration(t *testing.T) {

	if testing.Short() {
		t.Skip("Skipping integration test in short mode...")
	}

	devBaseDir := ".dev"

	// Create context
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

	// Configuration for module 1
	config := handler.NewConfiguration()
	config.Action = constants.Prepare
	config.BaseDir = baseDir
	config.Context = context
	config.ValidEventTypes = eventTypesStr
	config.PrintConfig = false
	config.LogLevel = "Debug"
	config.DevelopmentConfiguration = &development.Configuration{
		BaseDir: devBaseDir,
		Enabled: true,
	}

	environment := module.GetModuleEnvironment(baseDir)

	handler.Run(config)

	defer func() {
		_ = os.RemoveAll(baseDir)
		_ = os.RemoveAll(devBaseDir)
		_ = os.Remove(".memdb")
	}()

	// Check dev.prepared exists in development dir
	preparedPath := filepath.FromSlash(path.Join(config.DevelopmentConfiguration.ModuleDir, "prepared"))
	if _, err := os.Stat(preparedPath); os.IsNotExist(err) {
		t.Fatalf("prepared file should exist at path '%s'", preparedPath)
	}

	// Write an output image blob
	blob1 := "img1.png"
	blob1FilePath := filepath.FromSlash(path.Join(environment.OutputBlobDirPath, blob1))
	writeOutputBlob(blob1FilePath)

	// Write an output image blob
	blob2 := "subdir/img2.png"
	blob2FilePath := filepath.FromSlash(path.Join(environment.OutputBlobDirPath, blob2))
	writeOutputBlob(blob2FilePath)

	// Grab the length of the output directory
	var outFiles []string
	err := filepath.Walk(environment.OutputBlobDirPath, func(path string, f os.FileInfo, err error) error {
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
	handler.Run(config)

	// Check dev.committed exists in development dir
	committedPath := filepath.FromSlash(path.Join(config.DevelopmentConfiguration.ModuleDir, "committed"))
	if _, err := os.Stat(committedPath); os.IsNotExist(err) {
		t.Fatalf("committed file should exist at path '%s'", committedPath)
	}

	// Grab event ID from module 1's output event
	b, err := ioutil.ReadFile(filepath.FromSlash(path.Join(config.DevelopmentConfiguration.EventsDir(), "event0.json")))
	if err != nil {
		t.Fatalf("error reading event from disk '%+v'", err)
	}
	var inEvent common.Event
	err = json.Unmarshal(b, &inEvent)
	if err != nil {
		t.Fatalf("error unmarshalling event '%+v'", err)
	}

	// Set module 1 as the parent and set the input event ID
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
		t.Fatalf("error, input files length should match output length")
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

	// The first key, value pair should be as expected
	for _, kvp := range kvps {
		if kvp.Key != "abc" {
			t.Fatalf("expected key 'abc' in key value pairs: '%+v'", kvp)
		}
		if kvp.Value != "123" {
			t.Fatalf("expected key 'abc' to have value '123' in key value pairs: '%+v'", kvp)
		}
		break
	}
}

func writeOutputBlob(path string) error {
	dir := filepath.Dir(path)
	if dir != "." {
		_ = os.MkdirAll(dir, os.ModePerm)
	}
	err := ioutil.WriteFile(path, []byte("image"), os.ModePerm)
	if err != nil {
		return fmt.Errorf("error writing file '%s', '%+v'", path, err)
	}
	return nil
}

func writeOutputBytes(bytes []byte, path string) error {
	err := ioutil.WriteFile(path, bytes, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error writing file '%s', '%+v'", bytes, err)
	}
	return nil
}
