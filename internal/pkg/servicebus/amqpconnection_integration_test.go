package servicebus

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"pack.ag/amqp"

	"github.com/lawrencegripper/ion/internal/app/dispatcher/helpers"
	"github.com/lawrencegripper/ion/internal/pkg/messaging"
	"github.com/lawrencegripper/ion/internal/pkg/types"
)

func prettyPrintStruct(item interface{}) string {
	b, _ := json.MarshalIndent(item, "", " ")
	return string(b)
}

var config = &types.Configuration{
	ClientID:            os.Getenv("AZURE_CLIENT_ID"),
	ClientSecret:        os.Getenv("AZURE_CLIENT_SECRET"),
	ResourceGroup:       os.Getenv("AZURE_RESOURCE_GROUP"),
	SubscriptionID:      os.Getenv("AZURE_SUBSCRIPTION_ID"),
	TenantID:            os.Getenv("AZURE_TENANT_ID"),
	ServiceBusNamespace: os.Getenv("AZURE_SERVICEBUS_NAMESPACE"),
	Hostname:            "Test",
	SubscribesToEvent:   "ExampleEvent2",
	EventsPublished:     "ExamplePublishtopic",
	LogLevel:            "Debug",
	Job: &types.JobConfig{
		RetryCount: 5,
	},
}

// TestNewListener performs an end-2-end integration test on the listener talking to Azure ServiceBus
func TestIntegrationNewListener(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode...")
	}

	// pick a random name to prevent previous tests affecting this test
	config.ModuleName = helpers.RandomName(8)

	renewEvery := time.Second * 35
	processingTime := time.Second * 240

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*310)
	defer cancel()

	listener := NewAmqpConnection(ctx, config)
	// Remove topic to ensure each test has a clean topic to work with
	defer deleteSubscription(listener, config)

	nonce := time.Now().String()
	sender, err := listener.CreateAmqpSender(config.SubscribesToEvent)
	if err != nil {
		t.Error(err)
	}

	err = sender.Send(ctx, amqp.NewMessage([]byte(nonce)))
	if err != nil {
		t.Error(err)
	}

	stats, err := listener.GetQueueDepth()
	depth := stats.ActiveMessageCount
	if err != nil || depth == -1 {
		t.Error("Failed to get queue depth")
		t.Error(err)
	}

	if depth != 1 {
		t.Errorf("Expected queue depth of 1 Got:%v", depth)
		t.Fail()
	}

	amqpMessage, err := listener.Receiver.Receive(ctx)
	if err != nil {
		t.Error(err)
	}

	message := messaging.NewAmqpMessageWrapper(amqpMessage)

	// SUMMARY: Testing message lock renewal. By default SB messages's locks expire after 1min and the message is requeued
	// 1. Starts a loop renewing the message lock
	// 2. Block for more than 1min
	// 3. Accept the message (dequeuing it)
	// 4. Check the queue length is 0... if it's not we lost the lock and the message got put back on the queue.
	renewContext, cancel := context.WithCancel(ctx)
	go func() {
		for {
			select {
			case <-renewContext.Done():
				return
			default:
				time.Sleep(renewEvery)
				err := listener.RenewLocks(ctx, []*amqp.Message{
					amqpMessage,
				})

				// Report the test error if the context hasn't been cancelled.
				if err != nil {
					select {
					case <-renewContext.Done():
						return
					default:
						t.Error(err)
					}
				}
			}
		}
	}()

	time.Sleep(processingTime)
	cancel()

	err = message.Accept()
	if string(message.Body()) != nonce {
		t.Errorf("value not as expected in message Expected: %s Got: %s", nonce, message.Body())
	}

	// wait for the SB stats API to update
	time.Sleep(time.Second * 30)

	stats, err = listener.GetQueueDepth()
	depth = stats.ActiveMessageCount
	if err != nil || depth == -1 {
		t.Error("Failed to get queue depth")
		t.Error(err)
	}

	if depth != 0 {
		t.Errorf("Expected queue depth of 0 Got:%v", depth)
		t.Fail()
	}
}

func TestIntegrationRequeueReleasedMessages(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode...")
	}

	// pick a random name to prevent previous tests affecting this test
	config.ModuleName = helpers.RandomName(8)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()

	listener := NewAmqpConnection(ctx, config)
	// Remove topic to ensure each test has a clean topic to work with
	defer deleteSubscription(listener, config)

	nonce := time.Now().String()
	sender, err := listener.CreateAmqpSender(config.SubscribesToEvent)
	if err != nil {
		t.Error(err)
	}

	err = sender.Send(ctx, &amqp.Message{
		Value: nonce,
	})
	if err != nil {
		t.Error(err)
	}

	for index := 0; index < 6; index++ {
		amqpMessage, err := listener.Receiver.Receive(ctx)
		message := messaging.NewAmqpMessageWrapper(amqpMessage)
		if err != nil {
			t.Error(err)
		}

		if message.DeliveryCount() != index {
			t.Logf("Delivery count: Got %v Expected %v", message.DeliveryCount(), index)
		}

		err = message.Reject()
		if err != nil {
			t.Error(err)
		}
	}

	checkUntil := time.Now().Add(time.Second * 3)
	checkCtx, cancel := context.WithDeadline(context.Background(), checkUntil)
	defer cancel()

	_, err = listener.Receiver.Receive(checkCtx)
	if err != nil {
		t.Log(err)
	} else {
		t.Error("message delivered a 6th time - after 5 should be deadlettered")
	}
}

func deleteSubscription(listener *AmqpConnection, config *types.Configuration) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*45)
	defer cancel()
	_, err := listener.subsClient.Delete(ctx, config.ResourceGroup, config.ServiceBusNamespace, listener.TopicName, listener.SubscriptionName)
	if err != nil {
		panic(err)
	}
}
