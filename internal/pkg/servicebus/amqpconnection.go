package servicebus

import (
	"context"
	"fmt"
	"github.com/twinj/uuid"
	"net/http"
	"net/url"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2017-05-10/resources"
	"github.com/Azure/azure-sdk-for-go/services/servicebus/mgmt/2017-04-01/servicebus"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/lawrencegripper/ion/internal/app/dispatcher/helpers"
	"github.com/lawrencegripper/ion/internal/pkg/types"
	"pack.ag/amqp"
)

const serviceBusRootKeyName = "RootManageSharedAccessKey"

// AmqpConnection provides a connection to service bus and methods for creating required subscriptions and topics
type AmqpConnection struct {
	subsClient           *servicebus.SubscriptionsClient
	Endpoint             string
	SubscriptionName     string
	SubscriptionAmqpPath string
	TopicName            string
	AccessKeys           servicebus.AccessKeys
	AMQPConnectionString string
	Session              *amqp.Session
	Receiver             *amqp.Receiver
	ManagementReceiver   *amqp.Receiver
	ManagementSender     *amqp.Sender
	getSubscription      func() (servicebus.SBSubscription, error)
}

// MessageCountDetails is a mirror of the SB SDK object but without pointers and things we don't need so logrus can
// log the numbers correctly
type MessageCountDetails struct {
	// ActiveMessageCount - Number of active messages in the queue, topic, or subscription.
	ActiveMessageCount int64
	// DeadLetterMessageCount - Number of messages that are dead lettered.
	DeadLetterMessageCount int64
}

// GetQueueDepth returns the current length of the sb queue
func (l *AmqpConnection) GetQueueDepth() (MessageCountDetails, error) {
	sub, err := l.getSubscription()
	if err != nil || sub.MessageCount == nil {
		return MessageCountDetails{}, err
	}

	details := sub.CountDetails
	detailsPointerless := MessageCountDetails{}
	if details.ActiveMessageCount != nil {
		detailsPointerless.ActiveMessageCount = *details.ActiveMessageCount
	}
	if details.DeadLetterMessageCount != nil {
		detailsPointerless.DeadLetterMessageCount = *details.DeadLetterMessageCount
	}

	return detailsPointerless, nil
}

// Todo: Reconsider approach to error handling in this code.
// Move to returning err and panicing in the caller if listener creation fails.

// NewAmqpConnection initilises a servicebus lister from configuration
func NewAmqpConnection(ctx context.Context, config *types.Configuration) *AmqpConnection {
	if config == nil {
		log.Panic("Nil config not allowed")
	}
	if config.SubscribesToEvent == "" {
		log.Panic("Empty subscribesToEvent not allowed")
	}
	if config.ModuleName == "" {
		log.Panic("Empty module name not allowed")
	}
	if config.Job == nil {
		log.Panic("Job config required")
	}

	//Todo: close connection to amqp when context is cancelled/done

	listener := AmqpConnection{}
	auth := helpers.GetAzureADAuthorizer(config, azure.PublicCloud.ResourceManagerEndpoint)
	subsClient := servicebus.NewSubscriptionsClient(config.SubscriptionID)
	subsClient.Authorizer = auth
	topicsClient := servicebus.NewTopicsClient(config.SubscriptionID)
	topicsClient.Authorizer = auth
	namespaceClient := servicebus.NewNamespacesClient(config.SubscriptionID)
	namespaceClient.Authorizer = auth
	groupsClient := resources.NewGroupsClient(config.SubscriptionID)
	groupsClient.Authorizer = auth

	listener.subsClient = &subsClient

	// Check if resource group exists
	_, err := groupsClient.Get(ctx, config.ResourceGroup)
	if err != nil {
		log.WithField("config", types.RedactConfigSecrets(config)).Panicf("Failed getting resource group: %v", err)
	}

	// Check namespace exists
	namespace, err := namespaceClient.Get(ctx, config.ResourceGroup, config.ServiceBusNamespace)
	if err != nil {
		log.WithField("config", types.RedactConfigSecrets(config)).Panicf("Failed getting servicebus namespace: %v", err)
	}
	listener.Endpoint = *namespace.ServiceBusEndpoint

	keys, err := namespaceClient.ListKeys(ctx, config.ResourceGroup, config.ServiceBusNamespace, serviceBusRootKeyName)
	if err != nil {
		log.WithFields(log.Fields{
			"config":   types.RedactConfigSecrets(config),
			"response": keys,
		}).WithError(err).Panicf("Failed getting servicebus namespace")
	}

	listener.AccessKeys = keys
	listener.AMQPConnectionString = getAmqpConnectionString(*keys.KeyName, *keys.SecondaryKey, *namespace.Name)

	// Check Topic to listen on. Create a topic if missing
	topic := createTopic(ctx, topicsClient, config, config.SubscribesToEvent)
	listener.TopicName = strings.ToLower(*topic.Name)

	eventsPublished := strings.Split(config.EventsPublished, ",")
	for _, eventName := range eventsPublished {
		// Check topic to publish to. Create is missing
		createTopic(ctx, topicsClient, config, eventName)
	}

	// Check subscription to listen on. Create if missing
	subName := getSubscriptionName(config.SubscribesToEvent, config.ModuleName)
	sub, err := subsClient.Get(
		ctx,
		config.ResourceGroup,
		config.ServiceBusNamespace,
		config.SubscribesToEvent,
		subName,
	)
	listener.getSubscription = func() (servicebus.SBSubscription, error) {
		return subsClient.Get(
			ctx,
			config.ResourceGroup,
			config.ServiceBusNamespace,
			config.SubscribesToEvent,
			subName,
		)
	}

	if err != nil && sub.Response.StatusCode == http.StatusNotFound {
		log.WithField("config", types.RedactConfigSecrets(config)).Debugf("subscription %v doesn't exist.. creating", subName)

		deliveryCount := config.Job.RetryCount + 1
		if deliveryCount < 1 {
			log.Error("retryCount must be greater than or equal to 0")
		}

		subDef := servicebus.SBSubscription{
			SBSubscriptionProperties: &servicebus.SBSubscriptionProperties{
				MaxDeliveryCount: to.Int32Ptr(int32(deliveryCount)),
			},
		}
		sub, err = subsClient.CreateOrUpdate(
			ctx,
			config.ResourceGroup,
			config.ServiceBusNamespace,
			config.SubscribesToEvent,
			subName,
			subDef,
		)
		if err != nil {
			log.WithField("config", types.RedactConfigSecrets(config)).Panicf("Failed creating subscription: %v", err)
		}
	} else if err != nil {
		log.WithField("config", types.RedactConfigSecrets(config)).Panicf("Failed getting subscription: %v", err)
	}
	listener.SubscriptionName = *sub.Name
	listener.SubscriptionAmqpPath = getSubscriptionAmqpPath(config.SubscribesToEvent, config.ModuleName)

	listener.Session = createAmqpSession(&listener)
	listener.Receiver = createAmqpListener(&listener)
	listener.ManagementSender, listener.ManagementReceiver, err = listener.createAmqpSBManagementChannels(listener.TopicName, config.ModuleName)
	if err != nil {
		log.WithError(err).Error("failed to create management sender, without this renewal of message locks will fail")
	}

	return &listener
}

//RenewLocks renews the locks on messages provided
func (l *AmqpConnection) RenewLocks(ctx context.Context, messages []*amqp.Message) error {
	lockTokens := make([]amqp.UUID, 0, len(messages))
	for _, m := range messages {
		expires, ok := m.Annotations["x-opt-locked-until"]
		if !ok {
			log.WithField("message", m).Error("failed to get x-opt-locked-until from message annotations")
			fmt.Println("Failed getting locked unitl")
		} else {
			log.WithField("locked-until", expires).Info("message lock expires at")
			fmt.Println("GOT locked unitl")
			fmt.Printf("GOT locked until: %+v %+v \n", m.Annotations, m.DeliveryAnnotations)
		}

		lockToken, ok := m.DeliveryAnnotations["x-opt-lock-token"]
		if !ok {
			log.WithField("message", m).Error("failed to get x-opt-locktoken from message annotations, cannot renew lock")
			continue
		}
		lockTokenUUID, valid := lockToken.(amqp.UUID)
		if !valid {
			log.WithField("message", m).Error("failed to get x-opt-locktoken from message annotations - the type is not amqp.uuid, cannot renew lock")
			continue
		}

		lockTokens = append(lockTokens, lockTokenUUID)
		log.WithField("uuid", lockTokenUUID).Debug("adding lockid to renew")
	}

	if len(lockTokens) < 1 {
		log.Info("no lock tokens present to renew")
		return nil
	}

	hostname, _ := os.Hostname()
	err := l.ManagementSender.Send(ctx, &amqp.Message{
		ApplicationProperties: map[string]interface{}{
			"operation": "com.microsoft:renew-lock",
		},
		Properties: &amqp.MessageProperties{
			MessageID: uuid.NewV4().String(),
			ReplyTo:   hostname + "-receiver",
		},
		Value: map[string]interface{}{
			"lock-tokens": lockTokens,
		},
	})
	if err != nil {
		log.WithError(err).Error("failed to renew locks on active messages")
		return err
	}

	response, err := l.ManagementReceiver.Receive(ctx)
	if err != nil {
		// See https://github.com/lawrencegripper/ion/issues/157
		log.WithError(err).Error("error response from server on active messages, due to bug unmarshalling type 0x40 this is expected")
		return err
	}

	log.WithField("responseMsg", response).Debug("renew locks: response message")
	fmt.Printf("response message: %+v \n", response)
	fmt.Printf("properties %+v \n", response.Properties)

	return nil
}

func (l *AmqpConnection) createAmqpSBManagementChannels(topic, eventname string) (*amqp.Sender, *amqp.Receiver, error) {
	if l.Session == nil {
		log.WithField("currentListener", l).Panic("Cannot create amqp listener without a session already configured")
	}

	subscriptionAddress := getSubscriptionAmqpPath(topic, eventname) + "/$management"
	hostname, _ := os.Hostname()
	hostAddress := hostname
	// hostAddress := hostname + "/" + subscriptionAddress

	// receiver := uuid.NewV4().String()
	sender, err := l.Session.NewSender(
		amqp.LinkTargetAddress(subscriptionAddress),
		amqp.LinkSourceAddress(hostAddress+"-sender"),
	)

	if err != nil {
		log.Fatal("Creating management sender:", err)
		return nil, nil, err
	}

	reciever, err := l.Session.NewReceiver(
		amqp.LinkSourceAddress(subscriptionAddress),
		amqp.LinkTargetAddress(hostAddress+"-receiver"),
	)

	if err != nil {
		log.Fatal("Creating management receiver:", err)
		return nil, nil, err
	}

	return sender, reciever, nil
}

func createAmqpListener(listener *AmqpConnection) *amqp.Receiver {
	// Todo: how do we validate that the session is healthy?
	if listener.Session == nil {
		log.WithField("currentListener", listener).Panic("Cannot create amqp listener without a session already configured")
	}

	// Create a receiver
	receiver, err := listener.Session.NewReceiver(
		amqp.LinkSourceAddress(listener.SubscriptionAmqpPath),
		// amqp.LinkCredit(10), // Todo: Add config value to define how many inflight tasks the dispatcher can handle
	)
	if err != nil {
		log.Fatal("Creating receiver:", err)
	}

	return receiver
}

// CreateAmqpSender exists for e2e testing.
func (l *AmqpConnection) CreateAmqpSender(topic string) (*amqp.Sender, error) {
	if l.Session == nil {
		log.WithField("currentListener", l).Panic("Cannot create amqp listener without a session already configured")
	}

	sender, err := l.Session.NewSender(
		amqp.LinkTargetAddress("/" + topic),
	)
	if err != nil {
		log.Fatal("Creating receiver:", err)
		return nil, err
	}

	return sender, nil
}

func createTopic(ctx context.Context, topicsClient servicebus.TopicsClient, config *types.Configuration, topicName string) servicebus.SBTopic {
	topic, err := topicsClient.Get(ctx, config.ResourceGroup, config.ServiceBusNamespace, topicName)
	if err != nil && topic.Response.Response != nil && topic.Response.StatusCode == http.StatusNotFound {
		log.WithField("config", types.RedactConfigSecrets(config)).Debugf("topic %v doesn't exist.. creating", topicName)
		topic, err = topicsClient.CreateOrUpdate(ctx, config.ResourceGroup, config.ServiceBusNamespace, topicName, servicebus.SBTopic{})
		if err != nil {
			log.WithField("config", types.RedactConfigSecrets(config)).Panicf("Failed creating topic: %v", err)
		}
	} else if err != nil {
		log.WithField("config", types.RedactConfigSecrets(config)).Panicf("Failed getting topic: %v", err)
	}

	return topic
}

func createAmqpSession(listener *AmqpConnection) *amqp.Session {
	// Create client
	client, err := amqp.Dial(listener.AMQPConnectionString)
	if err != nil {
		log.Fatal("Dialing AMQP server:", err)
	}
	session, err := client.NewSession()
	if err != nil {
		log.WithError(err).Fatal("Creating session failed")
	}

	return session
}

func getAmqpConnectionString(keyName, keyValue, namespace string) string {
	encodedKey := url.QueryEscape(keyValue)
	return fmt.Sprintf("amqps://%s:%s@%s.servicebus.windows.net", keyName, encodedKey, namespace)
}

func getSubscriptionAmqpPath(eventName, moduleName string) string {
	return "/" + strings.ToLower(eventName) + "/subscriptions/" + getSubscriptionName(eventName, moduleName)
}

func getSubscriptionName(eventName, moduleName string) string {
	return strings.ToLower(eventName) + "_" + strings.ToLower(moduleName)
}
