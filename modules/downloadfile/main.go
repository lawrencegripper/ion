package main

import (
	"io"
	"net/http"
	"os"

	"github.com/lawrencegripper/ion/modules/helpers/Go/env"
	"github.com/lawrencegripper/ion/modules/helpers/Go/events"
	"github.com/lawrencegripper/ion/modules/helpers/Go/log"
)

func downloadFile(url, filepath string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close() //nolint: errcheck

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint: errcheck

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	env.MakeOutputDirs()

	filename := env.OutputDataDir() + "/file.raw"

	eventMeta, err := events.ReadEventMetaData()
	if err != nil {
		panic(err)
	}

	eventMetaMap := eventMeta.AsMap()
	if _, exist := eventMetaMap["url"]; !exist {
		log.Fatal("No link found")
		return
	}

	link := eventMetaMap["url"]

	if link == "" {
		log.Fatal("Empty link found")
		return
	}

	log.Debug("Downloading link: " + link)

	err = downloadFile(link, filename)
	if err != nil {
		log.Info(err.Error())
		return
	}

	events.Fire([]events.Event{
		{
			Event: "file_downloaded",
			File:  filename,
		},
	})
}
