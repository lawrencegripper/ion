package main

import (
	"github.com/lawrencegripper/ion/modules/helpers/Go/env"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestDownloaderModule(t *testing.T) {
	tempdir := path.Join(os.TempDir(), uuid.NewV4().String())
	os.MkdirAll(tempdir, 0777)

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
}
