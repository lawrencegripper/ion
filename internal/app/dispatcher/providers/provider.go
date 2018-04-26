package providers

import "github.com/lawrencegripper/ion/dispatcher/messaging"

// Provider is the interface a compute provider must fullfil to schedule and reconsile jobs
type Provider interface {
	Reconcile() error
	Dispatch(message messaging.Message) error
	InProgressCount() int
}
