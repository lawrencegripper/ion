package main

import (
	"encoding/json"
	"fmt"
	"github.com/lawrencegripper/ion/modules/helpers/Go/env"
	"github.com/lawrencegripper/ion/modules/helpers/Go/handler"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestDownloaderModule(t *testing.T) {
	tempdir := path.Join(os.TempDir(), uuid.NewV4().String())
	os.MkdirAll(tempdir, 0777)
	defer os.RemoveAll(tempdir)

	eventMetaFileBytes, err := ioutil.ReadFile("./testdata/eventmeta.json")
	if err != nil {
		t.FailNow()
	}
	os.Mkdir(path.Join(tempdir, "in"), 0777)
	err = ioutil.WriteFile(path.Join(tempdir, "in", "eventmeta.json"), eventMetaFileBytes, 0777)
	if err != nil {
		t.FailNow()
	}

	env.IonBaseDir = tempdir

	main()

	outDataFilePath := path.Join(tempdir, "out", "data")
	files, err := ioutil.ReadDir(outDataFilePath)
	if err != nil {
		t.Error("Failed getting files from out/data folder")
	}

	fileDownloaded := false
	for _, f := range files {
		if f.Name() == "file.raw" {
			downloadedFileBytes, err := ioutil.ReadFile(path.Join(outDataFilePath, f.Name()))
			if err != nil {
				t.Error("Failed to read in  downloaded file")
			}
			if string(downloadedFileBytes) != "Microsoft NCSI" {
				t.Errorf("Failed, download content not as expected got: %s", string(downloadedFileBytes))
			} else {
				fileDownloaded = true
			}
		}
	}

	if !fileDownloaded {
		t.Fail()
	}

	outEventsFilePath := path.Join(tempdir, "out", "events")
	eventFiles, _ := ioutil.ReadDir(outEventsFilePath)
	if len(eventFiles) != 1 {
		t.Error("Events not raised as expected")
		t.Error(eventFiles)
		return
	}

	expectedEvenFileData, _ := ioutil.ReadFile("./testdata/expectedEventFile.json")
	eventFileData, err := ioutil.ReadFile(path.Join(env.EventDir(), eventFiles[0].Name()))
	if err != nil {
		t.Error("Failed reading event file")
		t.Error(err)
	}

	if string(expectedEvenFileData) != string(eventFileData) {
		t.Error("Event file doesn't contain expected json")
		t.Log(string(eventFileData))
	}

	fmt.Println(path.Join(env.EventDir(), "out", "insights.json"))
	insightsFileData, err := ioutil.ReadFile(path.Join(env.IonBaseDir, "out", "insights.json"))
	insightsDeserialised := handler.Insights{}
	err = json.Unmarshal(insightsFileData, &insightsDeserialised)
	if err != nil {
		t.Error("Failed deserialising insights file")
		t.Error(err)
		t.Log(string(insightsFileData))
	}
	if insightsDeserialised[0].Key != "downloadTimeSec" {
		t.Error("Insights missing expected key")
		t.Log(insightsDeserialised)
	}

}
