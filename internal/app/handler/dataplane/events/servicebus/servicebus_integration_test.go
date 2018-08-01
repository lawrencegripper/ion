package servicebus

import (
	"context"
	"github.com/lawrencegripper/ion/internal/app/dispatcher/helpers"
	"github.com/lawrencegripper/ion/internal/pkg/common"
	sbamqp "github.com/lawrencegripper/ion/internal/pkg/servicebus"
	"github.com/lawrencegripper/ion/internal/pkg/types"
	"os"
	"pack.ag/amqp"
	"testing"
)

var config = &types.Configuration{
	ClientID:            os.Getenv("AZURE_CLIENT_ID"),
	ClientSecret:        os.Getenv("AZURE_CLIENT_SECRET"),
	ResourceGroup:       os.Getenv("AZURE_RESOURCE_GROUP"),
	SubscriptionID:      os.Getenv("AZURE_SUBSCRIPTION_ID"),
	TenantID:            os.Getenv("AZURE_TENANT_ID"),
	ServiceBusNamespace: os.Getenv("AZURE_SERVICEBUS_NAMESPACE"),
	Hostname:            "Test",
	ModuleName:          helpers.RandomName(8),
	SubscribesToEvent:   "exampleevent1235",
	EventsPublished:     "ExamplePublishtopic",
	LogLevel:            "Debug",
	Job: &types.JobConfig{
		RetryCount: 5,
	},
}

// TestNewListener performs an end-2-end integration with Service Bus sending from HTTP then receiving via AMQP
func TestIntegration_serivcebusHTTPSender(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode...")
	}

	ctx := context.Background()
	listenerm := sbamqp.NewAmqpConnection(ctx, config)
	defer listenerm.Receiver.Close(ctx)

	bus, err := NewServiceBus(&Config{
		AuthorizationRuleName: *listenerm.AccessKeys.KeyName,
		Enabled:               true,
		Key:                   *listenerm.AccessKeys.PrimaryKey,
		Topic:                 config.SubscribesToEvent,
		Namespace:             config.ServiceBusNamespace,
	})

	if err != nil {
		t.Fatal("failed to connect to sb")
	}

	err = bus.Publish(common.Event{
		Context: &common.Context{
			CorrelationID: "barrywhite",
		},
		Type: config.SubscribesToEvent,
	})

	if err != nil {
		t.Error(err)
		t.Fatal("failed to send")
	}

	msg, err := listenerm.Receiver.Receive(ctx)

	if err != nil {
		t.Error(err)
		t.Fatal("failed to receive")
	}

	if msg.DeliveryAnnotations == nil || len(msg.DeliveryAnnotations) < 1 {
		t.Errorf("expected delivery annotation to be set, have: %+v", msg)
		t.Logf("msg properties: %+v", msg.Properties)
		t.Log("msg annotation: %+v", msg.Annotations)
		t.Log("msg delivery annotation: %+v", msg.DeliveryAnnotations)
	}

	msg.Accept()

}

// TestNewListener performs an end-2-end integration with Service Bus sending from AMQP then receiving via AMQP
func TestIntegration_serivcebusAMQPSender(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode...")
	}

	ctx := context.Background()
	listenerm := sbamqp.NewAmqpConnection(ctx, config)
	defer listenerm.Receiver.Close(ctx)

	sender, err := listenerm.CreateAmqpSender(config.SubscribesToEvent)

	if err != nil {
		t.Error(err)
		t.Fatal("failed to send")
	}

	err = sender.Send(ctx, amqp.NewMessage([]byte("bob")))

	if err != nil {
		t.Error(err)
		t.Fatal("failed to send")
	}

	msg, err := listenerm.Receiver.Receive(ctx)

	if err != nil {
		t.Error(err)
		t.Fatal("failed to receive")
	}

	if msg.DeliveryAnnotations == nil || len(msg.DeliveryAnnotations) < 1 {
		t.Errorf("expected delivery annotation to be set, have: %+v", msg)
	}

	msg.Accept()

}
