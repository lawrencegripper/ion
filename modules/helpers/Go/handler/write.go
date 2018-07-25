package handler

import (
	"encoding/json"
	"github.com/lawrencegripper/ion/internal/pkg/common"
	"github.com/lawrencegripper/ion/modules/helpers/Go/env"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
	"strings"
)

//Event is an event to be raised by the module
type Event struct {
	Event string   `json:"event_type"`
	Files []string `json:"file"`
}

//Insights an array of keyValuePair
type Insights []Insight

//Insight wrapper for common.KeyValuePair
type Insight common.KeyValuePair

//WriteEvents creates an event file in the ion event dir which will be raise when
// the module exits with a zero exit code
// WARNING: this can only be called once per module execution.
func WriteEvents(events []Event) {
	i := 0
	for _, ev := range events {
		b, err := json.Marshal(common.KeyValuePairs{
			common.KeyValuePair{
				Key:   "eventType",
				Value: ev.Event,
			},
			common.KeyValuePair{
				Key:   "files",
				Value: strings.Join(ev.Files, ","),
			},
		})
		if err != nil {
			log.WithError(err).Panic("failed marshalling event")
		}
		f, err := os.Create(env.EventDir() + "/event-" + strconv.Itoa(i) + ".json")
		if err != nil {
			log.WithError(err).Panic("failed creating event file")
		}
		defer f.Close() //nolint: errcheck
		_, err = f.Write(b)
		if err != nil {
			log.WithError(err).Panic("failed writing to event file")
		}
		i = i + 1
	}
}

//WriteInsights creates an insights.json file with data to be stored by ion
// insights are stored in a searchable document store and usually contain
// information that could be queried. For example, names of items detected in the video.
// Insights will be stored only when the module exits with a zero exit code
// WARNING: this can only be called once per module execution.
func WriteInsights(i Insights) {
	b, err := json.Marshal(i)
	if err != nil {
		log.WithError(err).Panic("failed marshalling insights")
	}
	f, err := os.Create(env.InsightFile())
	if err != nil {
		log.WithError(err).Panic("failed creating insights file")
	}
	defer f.Close() //nolint: errcheck
	_, err = f.Write(b)
	if err != nil {
		log.WithError(err).Panic("failed writing to insights file")
	}
}
