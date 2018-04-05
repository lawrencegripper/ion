package events

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/lawrencegripper/ion/modules/helpers/Go/env"
)

type Event struct {
	Event string `json:"event_type"`
	File  string `json:"file"`
}

func Fire(events []Event) {
	i := 0
	for _, ev := range events {
		b, _ := json.Marshal(ev)
		f, _ := os.Create(env.EventDir + "/event-" + strconv.Itoa(i) + ".json")
		defer f.Close()
		f.Write(b)
		i = i + 1
	}
}
