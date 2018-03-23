package providers

import (
	"math/rand"
	"testing"
	"time"

	"github.com/lawrencegripper/mlops/dispatcher/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestNewListener performs an end-2-end integration test on the listener talking to Azure ServiceBus
func TestIntegrationKuberentesDispatch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode...")
	}

	config := &types.Configuration{
		Hostname:          mockDispatcherName,
		ModuleName:        "ModuleName",
		SubscribesToEvent: "ExampleEvent",
		LogLevel:          "Debug",
		JobConfig: &types.JobConfig{
			SidecarImage: "sidecarimagetest",
			WorkerImage:  "workerimagetest",
		},
	}

	p, err := NewKubernetesProvider(config)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	message := MockMessage{
		MessageID: randAlphaNumericSeq(12),
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
			jobsFoundCount++
		}
	}

	if jobsFoundCount != 1 {
		t.Error("Expected to only find 1 job")
	}

}

var lettersLower = []rune("abcdefghijklmnopqrstuvwxyz")

func randAlphaNumericSeq(n int) string {
	return randFromSelection(n, lettersLower)
}

func randFromSelection(length int, choices []rune) string {
	b := make([]rune, length)
	rand.Seed(time.Now().UnixNano())
	for i := range b {
		b[i] = choices[rand.Intn(len(choices))]
	}
	return string(b)
}
