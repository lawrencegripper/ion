package messaging

import (
	"encoding/json"
	"fmt"

	"github.com/lawrencegripper/ion/internal/pkg/common"

	log "github.com/sirupsen/logrus"
	"pack.ag/amqp"
)

// Message interface for any message protocol to use
type Message interface {
	ID() string
	DeliveryCount() int
	Body() []byte
	Accept() error
	Reject() error
	EventData() (common.Event, error)
	GetAMQPMessage() *amqp.Message
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

//GetAMQPMessage returns the wrapped message
func (m *AmqpMessage) GetAMQPMessage() *amqp.Message {
	return m.OriginalMessage
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
func (m *AmqpMessage) Body() []byte {
	return m.OriginalMessage.GetData()
}

// Accept mark the message as processed successfully (don't re-queue)
func (m *AmqpMessage) Accept() error {
	m.OriginalMessage.Accept()
	return nil
}

// Reject mark message to be requeued
func (m *AmqpMessage) Reject() error {
	m.OriginalMessage.Modify(true, false, nil)
	return nil
}

// EventData deserialize json value to type
func (m *AmqpMessage) EventData() (common.Event, error) {
	var event common.Event
	data := m.OriginalMessage.GetData()
	err := json.Unmarshal(data, &event)
	if err != nil {
		log.WithError(err).WithField("value", m.OriginalMessage.Data).Fatal("Unmarshal failed")
		return event, err
	}
	return event, nil
}
