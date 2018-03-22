package providers

import (
	"testing"

	"github.com/lawrencegripper/mlops/dispatcher/types"

	batchv1 "k8s.io/api/batch/v1"
)

const (
	mockDispatcherName = "mockdispatchername"
	mockMessageID      = "examplemessageID"
)

func NewMockKubernetesProvider(create func(b *batchv1.Job) (*batchv1.Job, error), list func() (*batchv1.JobList, error)) (*Kubernetes, error) {
	k := Kubernetes{}

	k.Namespace = "module-ns"
	k.jobConfig = &types.JobConfig{
		SidecarImage: "sidecar-image",
		WorkerImage:  "worker-image",
	}
	k.dispatcherName = mockDispatcherName

	k.inflightJobStore = map[string]Message{}
	k.createJob = create
	k.listAllJobs = list
	return &k, nil
}

func TestDispatchAddsJob(t *testing.T) {
	inMemMockJobStore := []batchv1.Job{}

	create := func(b *batchv1.Job) (*batchv1.Job, error) {
		inMemMockJobStore = append(inMemMockJobStore, *b)
		return b, nil
	}

	list := func() (*batchv1.JobList, error) {
		return &batchv1.JobList{
			Items: inMemMockJobStore,
		}, nil
	}

	k, _ := NewMockKubernetesProvider(create, list)

	messageToSend := MockMessage{
		MessageID: mockMessageID,
	}

	err := k.Dispatch(messageToSend)

	if err != nil {
		t.Error(err)
	}

	jobsLen := len(inMemMockJobStore)
	if jobsLen != 1 {
		t.Errorf("Job count incorrected Expected: 1 Got: %v", jobsLen)
	}
}

func TestDispatchedJobHasCorrectLabels(t *testing.T) {
	inMemMockJobStore := []batchv1.Job{}

	create := func(b *batchv1.Job) (*batchv1.Job, error) {
		inMemMockJobStore = append(inMemMockJobStore, *b)
		return b, nil
	}

	list := func() (*batchv1.JobList, error) {
		return &batchv1.JobList{
			Items: inMemMockJobStore,
		}, nil
	}

	k, _ := NewMockKubernetesProvider(create, list)

	messageToSend := MockMessage{
		MessageID: mockMessageID,
	}

	err := k.Dispatch(messageToSend)

	if err != nil {
		t.Error(err)
	}

	job := inMemMockJobStore[0]

	testCases := []struct {
		labelName     string
		expectedValue string
	}{
		{
			labelName:     dispatcherNameLabel,
			expectedValue: mockDispatcherName,
		},
		{
			labelName:     messageIDLabel,
			expectedValue: mockMessageID,
		},
		{
			labelName:     deliverycountlabel,
			expectedValue: "0",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run("Set:"+test.labelName, func(t *testing.T) {
			t.Parallel()

			value, ok := job.Labels[test.labelName]
			if !ok {
				t.Error("missing dispatchername label")
			}
			if value != test.expectedValue {
				t.Errorf("wrong dispatchername label Expected: %s Got: %s", test.expectedValue, value)
			}
		})
	}

}

func TestReconcileJobCompleted(t *testing.T) {
	//Setup... it's a long one. We need to schedule a job first
	inMemMockJobStore := []batchv1.Job{}

	create := func(b *batchv1.Job) (*batchv1.Job, error) {
		inMemMockJobStore = append(inMemMockJobStore, *b)
		return b, nil
	}

	list := func() (*batchv1.JobList, error) {
		return &batchv1.JobList{
			Items: inMemMockJobStore,
		}, nil
	}

	k, _ := NewMockKubernetesProvider(create, list)

	var acceptedMessage bool

	messageToSend := MockMessage{
		MessageID: mockMessageID,
		Accepted: func() {
			acceptedMessage = true
		},
	}

	err := k.Dispatch(messageToSend)
	if err != nil {
		t.Error(err)
	}

	job := &inMemMockJobStore[0]
	job.Status.Conditions = append(job.Status.Conditions, batchv1.JobCondition{
		Type: batchv1.JobComplete,
	})
	//Lets test things...
	err = k.Reconcile()
	if err != nil {
		t.Error(err)
	}

	if !acceptedMessage {
		t.Error("Failed to accept message during reconcilation. Expected message to be marked as accepted as job is complete")
	}
}

func TestReconcileJobFailed(t *testing.T) {
	//Setup... it's a long one. We need to schedule a job first
	inMemMockJobStore := []batchv1.Job{}

	create := func(b *batchv1.Job) (*batchv1.Job, error) {
		inMemMockJobStore = append(inMemMockJobStore, *b)
		return b, nil
	}

	list := func() (*batchv1.JobList, error) {
		return &batchv1.JobList{
			Items: inMemMockJobStore,
		}, nil
	}

	k, _ := NewMockKubernetesProvider(create, list)

	var rejectedMessage bool

	messageToSend := MockMessage{
		MessageID: mockMessageID,
		Rejected: func() {
			rejectedMessage = true
		},
	}

	err := k.Dispatch(messageToSend)
	if err != nil {
		t.Error(err)
	}

	job := &inMemMockJobStore[0]
	job.Status.Conditions = append(job.Status.Conditions, batchv1.JobCondition{
		Type: batchv1.JobFailed,
	})
	//Lets test things...
	err = k.Reconcile()
	if err != nil {
		t.Error(err)
	}

	if !rejectedMessage {
		t.Error("Failed to accept message during reconcilation. Expected message to be marked as accepted as job is complete")
	}
}

// AmqpMessage Wrapper for amqp
type MockMessage struct {
	MessageID string
	Accepted  func()
	Rejected  func()
}

// DeliveryCount get number of times the message has ben delivered
func (m MockMessage) DeliveryCount() int {
	return 0
}

// ID get the ID
func (m MockMessage) ID() string {
	// Todo: use reflection to identify type and do smarter stuff
	return m.MessageID
}

// Body get the body
func (m MockMessage) Body() interface{} {
	return "body"
}

// Accept mark the message as processed successfully (don't re-queue)
func (m MockMessage) Accept() error {
	m.Accepted()
	return nil
}

// Reject mark the message as failed and requeue
func (m MockMessage) Reject() error {
	m.Rejected()
	return nil
}
