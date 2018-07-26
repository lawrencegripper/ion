package providers

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"

	"github.com/Azure/azure-sdk-for-go/services/batch/2017-09-01.6.0/batch"
	"github.com/lawrencegripper/ion/internal/pkg/messaging"
	"github.com/lawrencegripper/ion/internal/pkg/types"
)

const (
	mockDispatcherName = "mockdispatchername"
	mockMessageID      = "examplemessageID"
)

func NewMockAzureBatchProvider(createTask func(taskDetails batch.TaskAddParameter) (autorest.Response, error), listTasks func() (*[]batch.CloudTask, error)) (*AzureBatch, error) {
	b := AzureBatch{}
	b.handlerArgs = []string{"--things=stuff", "--things=stuff", "--things=stuff", "--things=stuff", "--things=stuff"}
	b.jobConfig = &types.JobConfig{
		HandlerImage: "handler-image",
		WorkerImage:  "worker-image",
	}
	b.jobID = mockDispatcherName

	b.inprogressJobStore = map[string]messaging.Message{}
	b.createTask = createTask
	b.listTasks = listTasks
	b.removeTask = func(t *batch.CloudTask) (autorest.Response, error) {
		return autorest.Response{}, nil
	}
	b.logStore = &LogStore{}
	b.getLogs = func(*batch.CloudTask) string { return "logs" }
	return &b, nil
}

func TestAzureBatchDispatchAddsJob(t *testing.T) {
	inMemMockTaskStore := []batch.CloudTask{}

	create := func(taskDetails batch.TaskAddParameter) (autorest.Response, error) {
		inMemMockTaskStore = append(inMemMockTaskStore, batch.CloudTask{
			CommandLine: taskDetails.CommandLine,
		})
		return autorest.Response{}, nil
	}

	list := func() (*[]batch.CloudTask, error) {
		return &inMemMockTaskStore, nil
	}

	b, _ := NewMockAzureBatchProvider(create, list)

	messageToSend := MockMessage{
		MessageID: mockMessageID,
	}

	err := b.Dispatch(messageToSend)

	if err != nil {
		t.Error(err)
	}

	jobsLen := len(inMemMockTaskStore)
	if jobsLen != 1 {
		t.Errorf("Job count incorrected Expected: 1 Got: %v", jobsLen)
	}

	if inMemMockTaskStore[0].CommandLine == nil {
		t.Error("Command not passed to azure batch!")
	}
}

func TestAzureBatchDispatchAddsJob_CorrectPrepareAndCommit(t *testing.T) {
	inMemMockTaskStore := []batch.CloudTask{}

	create := func(taskDetails batch.TaskAddParameter) (autorest.Response, error) {
		inMemMockTaskStore = append(inMemMockTaskStore, batch.CloudTask{
			CommandLine: taskDetails.CommandLine,
		})
		return autorest.Response{}, nil
	}

	list := func() (*[]batch.CloudTask, error) {
		return &inMemMockTaskStore, nil
	}

	b, _ := NewMockAzureBatchProvider(create, list)

	messageToSend := MockMessage{
		MessageID: mockMessageID,
	}

	err := b.Dispatch(messageToSend)

	if err != nil {
		t.Error(err)
	}

	jobsLen := len(inMemMockTaskStore)
	if jobsLen != 1 {
		t.Errorf("Job count incorrected Expected: 1 Got: %v", jobsLen)
	}

	if inMemMockTaskStore[0].CommandLine == nil {
		t.Error("Command not passed to azure batch!")
	}

	t.Log(*inMemMockTaskStore[0].CommandLine)

	if strings.Count(*inMemMockTaskStore[0].CommandLine, "--action=prepare") != 1 {
		t.Error("Missing prepare action")
	}

	if strings.Count(*inMemMockTaskStore[0].CommandLine, "--action=commit") != 1 {
		t.Error("Missing commit action")
	}
}

func TestAzureBatchDispatchAddsJobWithGPU(t *testing.T) {
	//This is a very basic test.
	//Without mandating all dev/build machines have a gpu this is the best I can do
	inMemMockTaskStore := []batch.CloudTask{}

	create := func(taskDetails batch.TaskAddParameter) (autorest.Response, error) {
		inMemMockTaskStore = append(inMemMockTaskStore, batch.CloudTask{
			CommandLine: taskDetails.CommandLine,
		})
		return autorest.Response{}, nil
	}

	list := func() (*[]batch.CloudTask, error) {
		return &inMemMockTaskStore, nil
	}

	b, _ := NewMockAzureBatchProvider(create, list)

	b.batchConfig = &types.AzureBatchConfig{}
	b.batchConfig.RequiresGPU = true

	messageToSend := MockMessage{
		MessageID: mockMessageID,
	}

	err := b.Dispatch(messageToSend)

	if err != nil {
		t.Error(err)
	}

	jobsLen := len(inMemMockTaskStore)
	if jobsLen != 1 {
		t.Errorf("Job count incorrected Expected: 1 Got: %v", jobsLen)
		return
	}

	if inMemMockTaskStore[0].CommandLine == nil {
		t.Error("Command not passed to azure batch!")
	}

	resultTask := inMemMockTaskStore[0]
	if resultTask.CommandLine == nil {
		t.Error("commandline nil")
	}
	if !strings.Contains(*resultTask.CommandLine, "--runtime nvidia") {
		t.Error("Command missing nvidia runtime")
		t.Log(*resultTask.CommandLine)
	}
}

func TestAzureBatchFailedDispatchRejectsMessage(t *testing.T) {
	inMemMockTaskStore := []batch.CloudTask{}

	create := func(taskDetails batch.TaskAddParameter) (autorest.Response, error) {
		return autorest.Response{}, fmt.Errorf("Simulate error")
	}

	list := func() (*[]batch.CloudTask, error) {
		return &inMemMockTaskStore, nil
	}

	b, _ := NewMockAzureBatchProvider(create, list)

	wasRejected := false
	messageToSend := MockMessage{
		MessageID: mockMessageID,
	}
	messageToSend.Rejected = func() {
		wasRejected = true
	}

	err := b.Dispatch(messageToSend)

	if err == nil {
		t.Error("Expected error ... didn't see one!")
	}

	if !wasRejected {
		t.Error("Expected to be rejected... wasn't")
	}

	if len(inMemMockTaskStore) > 0 {
		t.Error("Expected job to not be stored")
	}
}

func TestAzureBatchReconcileJobCompleted(t *testing.T) {
	//Setup... it's a long one. We need to schedule a job first
	inMemMockTaskStore := []batch.CloudTask{}

	create := func(taskDetails batch.TaskAddParameter) (autorest.Response, error) {
		inMemMockTaskStore = append(inMemMockTaskStore, batch.CloudTask{
			ID:          to.StringPtr(mockMessageID),
			DisplayName: to.StringPtr(mockDispatcherName),
		})
		return autorest.Response{}, nil
	}

	list := func() (*[]batch.CloudTask, error) {
		return &inMemMockTaskStore, nil
	}

	b, _ := NewMockAzureBatchProvider(create, list)

	removeCalled := false
	b.removeTask = func(t *batch.CloudTask) (autorest.Response, error) {
		removeCalled = true
		return autorest.Response{}, nil
	}

	wasAccepted := false
	messageToSend := MockMessage{
		MessageID: mockMessageID,
	}
	messageToSend.Accepted = func() {
		wasAccepted = true
	}
	messageToSend.Rejected = func() {
		t.Error("Message rejected unexpectedly")
	}

	err := b.Dispatch(messageToSend)
	if err != nil {
		t.Error(err)
	}

	task := &inMemMockTaskStore[0]
	task.State = "completed"
	task.ExecutionInfo = &batch.TaskExecutionInformation{
		ExitCode: to.Int32Ptr(0),
	}
	//Lets test things...
	err = b.Reconcile()
	if err != nil {
		t.Error(err)
	}

	if !wasAccepted {
		t.Error("Failed to accept message during reconcilation. Expected message to be marked as accepted as job is complete")
	}

	if b.InProgressCount() != 0 {
		t.Error("Reconcile should remove jobs from the inmemory store once it has accepted or rejected them")
	}

	if !removeCalled {
		t.Error("Expected reconcile to remove completed task")
	}
}

func TestAzureBatchReconcileJobFailed(t *testing.T) {
	//Setup... it's a long one. We need to schedule a job first
	inMemMockTaskStore := []batch.CloudTask{}

	create := func(taskDetails batch.TaskAddParameter) (autorest.Response, error) {
		inMemMockTaskStore = append(inMemMockTaskStore, batch.CloudTask{
			ID: to.StringPtr(mockMessageID),
		})
		return autorest.Response{}, nil
	}

	list := func() (*[]batch.CloudTask, error) {
		return &inMemMockTaskStore, nil
	}

	b, _ := NewMockAzureBatchProvider(create, list)

	var rejectedMessage bool

	messageToSend := MockMessage{
		MessageID: mockMessageID,
		Rejected: func() {
			rejectedMessage = true
		},
	}

	err := b.Dispatch(messageToSend)
	if err != nil {
		t.Error(err)
	}

	task := &inMemMockTaskStore[0]
	task.State = "completed"
	task.ExecutionInfo = &batch.TaskExecutionInformation{
		ExitCode: to.Int32Ptr(127),
	}
	//Lets test things...
	err = b.Reconcile()
	if err != nil {
		t.Error(err)
	}

	if !rejectedMessage {
		t.Error("Failed to accept message during reconcilation. Expected message to be marked as accepted as job is complete")
	}

	if b.InProgressCount() != 0 {
		t.Error("Reconcile should remove jobs from the inmemory store once it has accepted or rejected them")
	}
}
