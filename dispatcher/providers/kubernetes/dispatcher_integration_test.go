package kubernetes

import (
	"encoding/json"
	"testing"
)

func prettyPrintStruct(item interface{}) string {
	b, _ := json.MarshalIndent(item, "", " ")
	return string(b)
}

// TestNewListener performs an end-2-end integration test on the listener talking to Azure ServiceBus
func TestIntegrationKuberentesDispatch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode...")
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Paniced: %v", prettyPrintStruct(r))
		}
	}()

	Dispatch(nil)
}
