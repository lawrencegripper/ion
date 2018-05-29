package providers

import (
	"strings"
	"testing"
	"time"

	"github.com/lawrencegripper/ion/internal/app/dispatcher/helpers"
	"github.com/lawrencegripper/ion/internal/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestNewListener performs an end-2-end integration test on the listener talking to Azure ServiceBus
func TestIntegrationKubernetesDispatch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode...")
	}

	config := &types.Configuration{
		Hostname:          mockDispatcherName,
		ModuleName:        "ModuleName",
		SubscribesToEvent: "ExampleEvent",
		Kubernetes: &types.KubernetesConfig{
			Namespace: "default",
		},
		LogLevel: "Debug",
		Job: &types.JobConfig{
			HandlerImage:       "handlerimagetest",
			WorkerImage:        "workerimagetest",
			MaxRunningTimeMins: 1,
		},
		Handler: &types.HandlerConfig{
			ServerPort:  1377,
			PrintConfig: true,
		},
	}

	p, err := NewKubernetesProvider(config, []string{"-examplearg1=1"})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	message := newNoOpMockMessage(helpers.RandomName(12))

	err = p.Dispatch(message)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	//Get scheduled jobs
	jobs, err := p.client.BatchV1().Jobs(p.Namespace).List(metav1.ListOptions{})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	jobsFoundCount := 0
	for _, j := range jobs.Items {
		if j.Name == getJobName(message) {
			CheckLabelsAssignedCorrectly(t, j, message.MessageID)
			handlerContainerArgs := j.Spec.Template.Spec.Containers[0].Args
			if len(handlerContainerArgs) > 3 {
				t.Logf("Some args set ... validate: %s", strings.Join(handlerContainerArgs, " "))
			}
			jobsFoundCount++
		}
	}

	if jobsFoundCount != 1 {
		t.Error("Expected to only find 1 job")
	}

}

func TestIntegrationKubernetesDispatch_MaxExecutionTime(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode...")
	}

	config := &types.Configuration{
		Hostname:          mockDispatcherName,
		ModuleName:        "ModuleName",
		SubscribesToEvent: "ExampleEvent",
		Kubernetes: &types.KubernetesConfig{
			Namespace: "default",
		},
		LogLevel: "Debug",
		Job: &types.JobConfig{
			HandlerImage:       "kubernetes/pause",
			WorkerImage:        "kubernetes/pause",
			MaxRunningTimeMins: 1,
		},
		Handler: &types.HandlerConfig{
			ServerPort:  1377,
			PrintConfig: true,
		},
	}

	p, err := NewKubernetesProvider(config, []string{"-examplearg1=1"})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	message := newNoOpMockMessage(helpers.RandomName(12))
	wasRejected := false
	message.Rejected = func() {
		wasRejected = true
	}

	err = p.Dispatch(message)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	time.Sleep(time.Minute * 2)

	err = p.Reconcile()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if !wasRejected {
		t.Error("Message wasn't rejected. Expected job to have timedout and been rejected")
	}
}
