package providers

import (
	"context"
	"fmt"
	"github.com/lawrencegripper/ion/dispatcher/helpers"
	"github.com/lawrencegripper/ion/dispatcher/types"
	log "github.com/sirupsen/logrus"
	"os"
	"testing"
	"time"
)

// TestNewListener performs an end-2-end integration test scheduling work onto Azure Batch
func TestIntegrationAzureBatchDispatch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode...")
	}

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
			SidecarImage: "sidecarimagetest",
			WorkerImage:  "workerimagetest",
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
	defer func() {
		_, err := p.jobClient.Delete(p.ctx, p.dispatcherName, nil, nil, nil, nil, "", "", nil, nil)
		if err != nil {
			fmt.Println(err)
		}
	}()

	messageAccepted := false
	messageRejected := false
	message := MockMessage{
		MessageID: helpers.RandomName(12),
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
	waitCtx, cancel := context.WithTimeout(ctx, time.Minute*4)
	defer cancel()

	for {
		if _, ok := waitCtx.Deadline(); !ok {
			t.Log("Timedout")
			break
		}

		log.Info("Checking for completed task")

		//Get scheduled jobs
		tasksPtr, err := p.listTasks()
		if err != nil {
			t.Log("Failed retreiving tasks from Azure Batch")
			t.Error(err)
		}
		if tasksPtr == nil {
			t.Log("Failed retreiving tasks from Azure Batch - tasks nil")
			t.Fail()
		}
		tasks := *tasksPtr

		if len(tasks) != 1 {
			t.Error("Expected to only find 1 job")
		}
		log.WithField("tasks", tasks).Info("Found tasks...")
		for _, task := range tasks {
			if task.ExecutionInfo == nil || task.ExecutionInfo.ExitCode == nil {
				continue
			}
			exitCode := *task.ExecutionInfo.ExitCode
			if exitCode != 0 {
				t.Error("The task failed to execute in Azure batch")
				t.Fail()
				return
			} else if exitCode == 0 {
				t.Log("Task completed with zero exit code - happy days!")
				return
			}

		}
		time.Sleep(time.Second * 5)
	}

	if !messageAccepted {
		t.Error("Expected message to be accepted")
	}

	if messageRejected {
		t.Error("Message rejected - was expecting it to be accepted")
	}
}
