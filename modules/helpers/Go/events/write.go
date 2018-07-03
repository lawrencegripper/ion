package events

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/lawrencegripper/ion/modules/helpers/Go/env"
)

//Event is an event to be raised by the module
type Event struct {
	Event string `json:"event_type"`
	File  string `json:"file"`
}

//Fire creates an event file in the ion event dir which will be raise when
// the module exits with a non-zero exit code
func Fire(events []Event) {
	i := 0
	for _, ev := range events {
		b, err := json.Marshal(ev)
		if err != nil {
			panic("failed marshalling event")
		}
		f, err := os.Create(env.EventDir() + "/event-" + strconv.Itoa(i) + ".json")
		if err != nil {
			panic("failed creating event file")
		}
		defer f.Close() //nolint: errcheck
		_, err = f.Write(b)
		if err != nil {
			panic("failed writing to event file")
		}
		i = i + 1
	}
}
