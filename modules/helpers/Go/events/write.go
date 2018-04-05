package events

import (
	"encoding/json"
	"os"
	"strconv"
)

type Event struct {
	Event string `json:"event_type"`
	File  string `json:"file"`
}

const (
	EventDir = "out/events"
)

func Fire(events []Event) {
	i := 0
	for _, ev := range events {
		b, _ := json.Marshal(ev)
		f, _ := os.Create(EventDir + "/event-" + strconv.Itoa(i) + ".json")
		defer f.Close()
		f.Write(b)
		i = i + 1
	}
}
