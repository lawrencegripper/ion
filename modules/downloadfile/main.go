package main

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/lawrencegripper/ion/modules/helpers/Go/env"
	"github.com/lawrencegripper/ion/modules/helpers/Go/events"
	"github.com/lawrencegripper/ion/modules/helpers/Go/log"
	"github.com/lawrencegripper/ion/modules/helpers/Go/handler"
)

func downloadFile(url, filepath string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	handler.Ready()

	filename := env.OutputDataDir + "/file.raw"

	dat, _ := ioutil.ReadFile(env.InputDataDir + "/link.txt")
	link := string(dat)

	if link == "" {
		log.Fatal("No link found")
		return
	}

	log.Debug("Downloading link: " + link)

	err := downloadFile(link, filename)
	if err != nil {
		log.Info(err.Error())
		return
	}

	handler.Commit()

	events.Fire([]events.Event{
		events.Event{
			Event: "file_downloaded",
			File:  filename,
		},
	})
}
