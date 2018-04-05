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

//MockEventPublisher is a mock event publisher implementation
type MockEventPublisher struct {
	baseDir string
	count   int
}

//NewMockEventPublisher returns a new MockEventPublisher object
func NewMockEventPublisher(dir string) *MockEventPublisher {
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return nil
	}
	return &MockEventPublisher{
		baseDir: dir,
		count:   0,
	}
}

//Publish is a mock implementation of the Publish method
func (e *MockEventPublisher) Publish(event types.Event) error {
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
func (e *MockEventPublisher) Close() {
	_ = os.RemoveAll(e.baseDir)
}
