package providers

import (
	"github.com/lawrencegripper/ion/dispatcher/messaging"
)

//Check providers match interface at compile time
var _ Provider = &AzureBatch{}

// AzureBatch schedules jobs onto k8s from the queue and monitors their progress
type AzureBatch struct {
	inprogressJobStore map[string]messaging.Message
	dispatcherName     string
	sidecarArgs        []string
	sidecarEnvVars     map[string]interface{}
}

// InProgressCount will show how many tasks are currently in progress
func (batch *AzureBatch) InProgressCount() int {
	panic("Not Implimented")
}

// Dispatch will dispatch a job onto Azure Batch
func (batch *AzureBatch) Dispatch(message messaging.Message) error {
	panic("Not Implimented")
}

// Reconcile will check inprogress tasks against and accept/reject messages were the job has completed/failed
func (batch *AzureBatch) Reconcile() error {
	panic("Not Implimented")
}
