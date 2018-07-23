package providers

import (
	"github.com/lawrencegripper/ion/internal/pkg/messaging"
	log "github.com/sirupsen/logrus"
)

// Provider is the interface a compute provider must fullfil to schedule and reconsile jobs
type Provider interface {
	Reconcile() error
	Dispatch(message messaging.Message) error
	InProgressCount() int
	GetActiveMessages() []messaging.Message
}

//GetLoggerForMessage Adds context fields to the logger for the message
func GetLoggerForMessage(message messaging.Message, l *log.Entry) *log.Entry {
	if message == nil {
		return l.WithField("inputError", "nil message provided to 'getloggerformessage' func")
	}
	return l.WithField("messagebody", string(message.Body())).WithField("messageID", message.ID())
}
