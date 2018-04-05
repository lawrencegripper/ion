package mock

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"

	"github.com/lawrencegripper/ion/sidecar/types"
)

//EventPublisher is a mock event publisher implementation
type EventPublisher struct {
	baseDir string
	count   int
}

//NewEventPublisher returns a new EventPublisher object
func NewEventPublisher(dir string) *EventPublisher {
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return nil
	}
	return &EventPublisher{
		baseDir: dir,
		count:   0,
	}
}

//Publish is a mock implementation of the Publish method
func (e *EventPublisher) Publish(event types.Event) error {
	eventPath := path.Join(e.baseDir, "event"+strconv.Itoa(e.count)+".json")
	eventJSON, err := json.Marshal(&event)
	if err != nil {
		return fmt.Errorf("error marshalling event '%+v'", err)
	}
	err = ioutil.WriteFile(eventPath, eventJSON, 0777)
	if err != nil {
		return fmt.Errorf("error writing event to file '%s': '%+v'", eventPath, err)
	}
	return nil
}

//Close is a mock implementation of the Close method
func (e *EventPublisher) Close() {
	_ = os.RemoveAll(e.baseDir)
}
