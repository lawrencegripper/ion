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

func TestBasicTranscode(t *testing.T) {
	fmt.Println("WARNING: This test requires FFMPEG to be installed and on the PATH")

	tempdir := path.Join(os.TempDir(), uuid.NewV4().String())
	os.MkdirAll(tempdir, 0777)
	defer os.RemoveAll(tempdir)

	eventMetaFileBytes, err := ioutil.ReadFile("./testdata/bird.avi")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	os.MkdirAll(path.Join(tempdir, "in", "data"), 0777)
	err = ioutil.WriteFile(path.Join(tempdir, "in", "data", "file.raw"), eventMetaFileBytes, 0777)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	os.Setenv("FFMPEG_USE_GPU", "false")
	env.IonBaseDir = tempdir

	main()

	outDataFilePath := path.Join(tempdir, "out", "data")
	files, err := ioutil.ReadDir(outDataFilePath)
	if err != nil {
		t.Error("Failed getting files from out/data folder")
	}

	fileDownloaded := false
	for _, f := range files {
		if f.Name() == "file.raw-1280x720-h264.mp4" {
			fileDownloaded = true
		}
	}

	if !fileDownloaded {
		t.Error("transcoded file missing")
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
	if insightsDeserialised[0].Key != "transcodeTimeSec" {
		t.Error("Insights missing expected key")
		t.Log(insightsDeserialised)
	}
}
