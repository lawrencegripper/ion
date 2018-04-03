package mock

import "github.com/lawrencegripper/ion/sidecar/types"

//MockEventPublisher is a mock event publisher implementation
type MockEventPublisher struct {
}

//NewMockEventPublisher returns a new MockEventPublisher object
func NewMockEventPublisher() *MockEventPublisher {
	return &MockEventPublisher{}
}

//Publish is a mock implementation of the Publish method
func (e *MockEventPublisher) Publish(event types.Event) error {
	return nil
}

//Close is a mock implementation of the Close method
func (e *MockEventPublisher) Close() {
}
