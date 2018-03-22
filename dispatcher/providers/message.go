package providers

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"pack.ag/amqp"
)

// Message interface for any message protocol to use
type Message interface {
	ID() string
	DeliveryCount() int
	Body() interface{}
	Accept() error
	Reject() error
}

// AmqpMessage Wrapper for amqp
type AmqpMessage struct {
	// Todo: Should this be private?
	OriginalMessage *amqp.Message
}

// NewAmqpMessageWrapper get number of times the message has ben delivered
func NewAmqpMessageWrapper(m *amqp.Message) Message {
	if m == nil {
		log.Panic("Message cannot be nil")
	}
	return &AmqpMessage{
		OriginalMessage: m,
	}
}

// DeliveryCount get number of times the message has ben delivered
func (m *AmqpMessage) DeliveryCount() int {
	return int(m.OriginalMessage.Header.DeliveryCount)
}

// ID get the ID
func (m *AmqpMessage) ID() string {
	// Todo: use reflection to identify type and do smarter stuff
	return fmt.Sprintf("%v", m.OriginalMessage.Properties.MessageID)
}

// Body get the body
func (m *AmqpMessage) Body() interface{} {
	return m.OriginalMessage.Value
}

// Accept mark the message as processed successfully (don't re-queue)
func (m *AmqpMessage) Accept() error {
	m.OriginalMessage.Accept()
	return nil
}

// Reject mark the message as failed and requeue
func (m *AmqpMessage) Reject() error {
	m.OriginalMessage.Reject()
	return nil
}
