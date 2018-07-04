package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/lawrencegripper/ion/modules/helpers/Go/env"
	"github.com/lawrencegripper/ion/modules/helpers/Go/handler"
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

	downloadedFileName := "file.raw"
	downloadedFilePath := path.Join(env.OutputDataDir(), downloadedFileName)

	eventMeta, err := handler.ReadEventMetaData()
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

	start := time.Now()

	err = downloadFile(link, downloadedFilePath)
	elapsed := time.Since(start)
	if err != nil {
		log.Info(err.Error())
		return
	}

	handler.WriteInsights(handler.Insights{
		handler.Insight{
			Key:   "downloadTimeSec",
			Value: fmt.Sprintf("%.6f", elapsed.Seconds()),
		},
	})

	handler.WriteEvents([]handler.Event{
		{
			Event: "file_downloaded",
			File:  downloadedFileName,
		},
	})
}
