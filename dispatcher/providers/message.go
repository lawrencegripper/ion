package providers

import (
	"pack.ag/amqp"
)

// Message interface for any message protocol to use
type Message interface {
	ID() string
	DeliveryCount() uint32
	Body() interface{}
	Accept() error
	Reject() error
}

// AmqpMessage Wrapper for amqp
type AmqpMessage struct {
	OriginalMessage amqp.Message
}

// DeliveryCount get number of times the message has ben delivered
func (m *AmqpMessage) DeliveryCount() uint32 {
	return m.OriginalMessage.Header.DeliveryCount
}

// ID get the ID
func (m *AmqpMessage) ID() string {
	return string(m.OriginalMessage.Properties.MessageID)
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
