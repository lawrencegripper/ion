package providers

import (
	"testing"

	"github.com/lawrencegripper/mlops/dispatcher/types"
)

// TestNewListener performs an end-2-end integration test on the listener talking to Azure ServiceBus
func TestIntegrationKuberentesDispatch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode...")
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Paniced: %v", types.PrettyPrintStruct(r))
		}
	}()

	config := &types.Configuration{
		Hostname:          "Test",
		ModuleName:        "ModuleName",
		SubscribesToEvent: "ExampleEvent",
		LogLevel:          "Debug",
	}

	p, err := NewKubernetesProvider(config)
	if err != nil {
		t.Error(err)
	}

	p.Dispatch(nil)
}
