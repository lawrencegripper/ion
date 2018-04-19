package providers

import (
	"strings"
	"testing"

	"github.com/lawrencegripper/ion/dispatcher/helpers"
	"github.com/lawrencegripper/ion/dispatcher/types"
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
			Namespace: "integrationtesting-ion",
		},
		LogLevel: "Debug",
		Job: &types.JobConfig{
			SidecarImage: "sidecarimagetest",
			WorkerImage:  "workerimagetest",
		},
		Sidecar: &types.SidecarConfig{
			ServerPort:  1377,
			PrintConfig: true,
		},
	}

	p, err := NewKubernetesProvider(config, []string{"-examplearg1=1"})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	message := MockMessage{
		MessageID: helpers.RandomName(12),
	}

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
			sidecarContainerArgs := j.Spec.Template.Spec.Containers[0].Args
			if len(sidecarContainerArgs) > 3 {
				t.Logf("Some args set ... validate: %s", strings.Join(sidecarContainerArgs, " "))
			}
			jobsFoundCount++
		}
	}

	if jobsFoundCount != 1 {
		t.Error("Expected to only find 1 job")
	}

}
