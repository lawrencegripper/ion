package servicebus

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Azure/go-autorest/autorest/to"

	log "github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2017-05-10/resources"
	"github.com/Azure/azure-sdk-for-go/services/servicebus/mgmt/2017-04-01/servicebus"
	"github.com/lawrencegripper/mlops/dispatcher/types"
	"pack.ag/amqp"
)

const serviceBusRootKeyName = "RootManageSharedAccessKey"

// Listener provides a connection to service bus and methods for creating required subscriptions and topics
type Listener struct {
	subsClient           *servicebus.SubscriptionsClient
	Endpoint             string
	SubscriptionName     string
	SubscriptionAmqpPath string
	TopicName            string
	AccessKeys           servicebus.AccessKeys
	AMQPConnectionString string
	AmqpSession          *amqp.Session
	AmqpReceiver         *amqp.Receiver
	getSubscription      func() (servicebus.SBSubscription, error)
}

// GetQueueDepth returns the current length of the sb queue
func (l *Listener) GetQueueDepth() (*int64, error) {
	sub, err := l.getSubscription()
	if err != nil {
		return nil, err
	}

	return sub.MessageCount, nil
}

// Todo: Reconsider approach to error handling in this code.
// Move to returning err and panicing in the caller if listener creation fails.

// NewListener initilises a servicebus lister from configuration
func NewListener(ctx context.Context, config *types.Configuration) *Listener {
	if config == nil {
		log.Panic("Nil config not allowed")
	}

	//Todo: close connection to amqp when context is cancelled/done

	listener := Listener{}
	auth := getAuthorizer(config)
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

	// Check topic to publish to. Create is missing
	createTopic(ctx, topicsClient, config, config.EventsPublished)

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
		subDef := servicebus.SBSubscription{
			SBSubscriptionProperties: &servicebus.SBSubscriptionProperties{
				MaxDeliveryCount: to.Int32Ptr(int32(config.Job.RetryCount)),
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

	listener.AmqpSession = createAmqpSession(&listener)
	listener.AmqpReceiver = createAmqpListener(&listener)

	return &listener
}

func createAmqpListener(listener *Listener) *amqp.Receiver {
	// Todo: how do we validate that the session is healthy?
	if listener.AmqpSession == nil {
		log.WithField("currentListener", listener).Panic("Cannot create amqp listener without a session already configured")
	}

	// Create a receiver
	receiver, err := listener.AmqpSession.NewReceiver(
		amqp.LinkSourceAddress(listener.SubscriptionAmqpPath),
		// amqp.LinkCredit(10), // Todo: Add config value to define how many inflight tasks the dispatcher can handle
	)
	if err != nil {
		log.Fatal("Creating receiver:", err)
	}

	return receiver
}

func createTopic(ctx context.Context, topicsClient servicebus.TopicsClient, config *types.Configuration, topicName string) servicebus.SBTopic {
	topic, err := topicsClient.Get(ctx, config.ResourceGroup, config.ServiceBusNamespace, topicName)
	if err != nil && topic.Response.StatusCode == http.StatusNotFound {
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

func createAmqpSession(listener *Listener) *amqp.Session {
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
