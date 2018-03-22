package servicebus

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"pack.ag/amqp"

	"github.com/lawrencegripper/mlops/dispatcher/types"
	log "github.com/sirupsen/logrus"
)

func prettyPrintStruct(item interface{}) string {
	b, _ := json.MarshalIndent(item, "", " ")
	return string(b)
}

// TestNewListener performs an end-2-end integration test on the listener talking to Azure ServiceBus
func TestIntegrationNewListener(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode...")
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Paniced: %v", prettyPrintStruct(r))
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	listener := NewListener(ctx, &types.Configuration{
		ClientID:            os.Getenv("AZURE_CLIENT_ID"),
		ClientSecret:        os.Getenv("AZURE_CLIENT_SECRET"),
		ResourceGroup:       os.Getenv("AZURE_RESOURCE_GROUP"),
		SubscriptionID:      os.Getenv("AZURE_SUBSCRIPTION_ID"),
		TenantID:            os.Getenv("AZURE_TENANT_ID"),
		ServiceBusNamespace: os.Getenv("AZURE_SERVICEBUS_NAMESPACE"),
		Hostname:            "Test",
		ModuleName:          "ModuleName",
		SubscribesToEvent:   "ExampleEvent",
		LogLevel:            "Debug",
	})

	nonce := time.Now().String()
	sender := createAmqpSender(listener)
	err := sender.Send(ctx, &amqp.Message{
		Value: nonce,
	})
	if err != nil {
		t.Error(err)
	}

	message, err := listener.AmqpReceiver.Receive(ctx)
	if err != nil {
		t.Error(err)
	}

	message.Accept()
	if message.Value != nonce {
		t.Errorf("value not as expected in message Expected: %s Got: %s", nonce, message.Value)
	}

	depth, err := listener.GetQueueDepth()
	if err != nil || depth == nil {
		t.Error("Failed to get queue depth")
		t.Error(err)
	}

	derefDepth := *depth

	if derefDepth != 0 {
		t.Errorf("Expected queue depth of 0 Got:%v", derefDepth)
		t.Fail()
	}
}

// createAmqpSender exists for e2e testing.
func createAmqpSender(listener *Listener) *amqp.Sender {
	if listener.AmqpSession == nil {
		log.WithField("currentListener", listener).Panic("Cannot create amqp listener without a session already configured")
	}

	sender, err := listener.AmqpSession.NewSender(
		amqp.LinkTargetAddress("/" + listener.TopicName),
	)
	if err != nil {
		log.Fatal("Creating receiver:", err)
	}

	return sender
}
