package messaging

import (
	"encoding/json"
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
	EventData() (Event, error)
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

// Reject mark message to be requeued
func (m *AmqpMessage) Reject() error {
	// Todo: fix this!
	log.Error("WARNING: REJECTED message doesn't correctly increment delivery count")
	log.Error("WARNING: REJECTED message doesn't correctly increment delivery count")
	log.Error("WARNING: REJECTED message doesn't correctly increment delivery count")
	log.Error("WARNING: REJECTED message doesn't correctly increment delivery count")
	log.Error("WARNING: REJECTED message doesn't correctly increment delivery count")
	log.Error("WARNING: REJECTED message doesn't correctly increment delivery count")
	m.OriginalMessage.Release()
	return nil
}

// EventData deserialize json value to type
func (m *AmqpMessage) EventData() (Event, error) {
	a := Event{}
	err := json.Unmarshal([]byte(fmt.Sprintf("%v", m.OriginalMessage.Value)), &a)
	if err != nil {
		log.WithError(err).WithField("value", m.OriginalMessage.Value).Fatal("Unmarshal failed")
		return a, err
	}
	return a, nil
}
