package providers

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/lawrencegripper/ion/internal/app/dispatcher/helpers"
	"github.com/lawrencegripper/ion/internal/pkg/types"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.FatalLevel)
}

// TestIntegrationAzureBatchDispatch performs an end-2-end integration test scheduling work onto Azure Batch
func TestIntegrationAzureBatchDispatch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode...")
	}

	testCases := []struct {
		name                  string
		dockerimage           string
		maxExecutionTimeMins  int
		expectedExitCode      int32
		expectMessageAccepted bool
		expectMessageRejected bool
	}{
		{
			name:                  "successful_module",
			dockerimage:           "lawrencegripper/busyboxecho",
			expectedExitCode:      0,
			expectMessageAccepted: true,
			maxExecutionTimeMins:  5,
		},
		{
			name:                  "exceedMaxExecTime_module",
			dockerimage:           "kubernetes/pause",
			expectedExitCode:      2,
			expectMessageRejected: true,
			maxExecutionTimeMins:  1,
		},
		{
			name:                  "failing_module",
			dockerimage:           "imagethatdoesntexist",
			expectedExitCode:      125,
			expectMessageRejected: true,
			maxExecutionTimeMins:  5,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run("Set:"+test.name, func(t *testing.T) {
			config := &types.Configuration{
				Hostname:          mockDispatcherName,
				ModuleName:        "ModuleName",
				SubscribesToEvent: "ExampleEvent",
				LogLevel:          "Debug",
				ClientID:          os.Getenv("AZURE_CLIENT_ID"),
				ClientSecret:      os.Getenv("AZURE_CLIENT_SECRET"),
				ResourceGroup:     os.Getenv("AZURE_RESOURCE_GROUP"),
				SubscriptionID:    os.Getenv("AZURE_SUBSCRIPTION_ID"),
				TenantID:          os.Getenv("AZURE_TENANT_ID"),
				Job: &types.JobConfig{
					SidecarImage:       test.dockerimage,
					WorkerImage:        test.dockerimage,
					MaxRunningTimeMins: test.maxExecutionTimeMins,
				},
				AzureBatch: &types.AzureBatchConfig{
					BatchAccountLocation: os.Getenv("AZURE_BATCH_ACCOUNT_LOCATION"),
					BatchAccountName:     os.Getenv("AZURE_BATCH_ACCOUNT_NAME"),
					JobID:                helpers.RandomName(12),
					PoolID:               "testpool",
				},
			}

			p, err := NewAzureBatchProvider(config, []string{"-examplearg1=1"})
			if err != nil {
				t.Error(err)
				t.FailNow()
			}

			messageAccepted := false
			messageRejected := false
			message := MockMessage{
				MessageID: helpers.RandomName(6),
			}
			message.Accepted = func() {
				messageAccepted = true
			}
			message.Rejected = func() {
				messageRejected = true
			}

			err = p.Dispatch(message)
			if err != nil {
				t.Error(err)
				t.FailNow()
			}

			ctx := context.Background()
			waitCtx, cancel := context.WithTimeout(ctx, time.Minute*2)
			defer cancel()

		loop:
			for {
				//Get task
				task, err := p.taskClient.Get(p.ctx, p.dispatcherName, message.ID(), "", "", nil, nil, nil, nil, "", "", nil, nil)
				if err != nil {
					t.Error(err)
				}

				select {
				case <-waitCtx.Done():
					t.Log("Timed-out")
					break loop
				default:
					t.Log("Checking Task")
				}

				log.Info("Checking for completed task")

				if task.ExecutionInfo == nil || task.ExecutionInfo.ExitCode == nil || *task.ID != message.ID() || task.State != "completed" {
					time.Sleep(time.Second * 5)
					continue
				}

				exitCode := *task.ExecutionInfo.ExitCode
				t.Log(exitCode)

				if exitCode == test.expectedExitCode {
					t.Logf("Success - Error code: %v expected: %v", exitCode, test.expectedExitCode)
				} else {
					t.Errorf("Error code: %v expected: %v", exitCode, test.expectedExitCode)
				}
				break
			}

			err = p.Reconcile()
			if err != nil {
				t.Error(err)
			}

			if test.expectMessageAccepted && !messageAccepted {
				t.Error("Message wasn't accepted")
			}

			if test.expectMessageRejected && !messageRejected {
				t.Error("Message wasn't rejected")
			}
		})

	}

}
