package servicebus

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/azure/azure-sdk-for-go/services/resources/mgmt/2017-05-10/resources"
	"github.com/azure/azure-sdk-for-go/services/servicebus/mgmt/2017-04-01/servicebus"
	"github.com/lawrencegripper/mlops/dispatcher/types"
	"pack.ag/amqp"
)

const serviceBusRootKeyName = "RootManageSharedAccessKey"

// Listener provides a connection to service bus and methods for creating required subscriptions and topics
type Listener struct {
	namespaceClient      servicebus.NamespacesClient
	subscriptionsClient  servicebus.SubscriptionsClient
	topicsClient         servicebus.TopicsClient
	Endpoint             string
	SubscriptionName     string
	SubscriptionAmqpPath string
	TopicName            string
	AccessKeys           servicebus.AccessKeys
	AMQPConnectionString string
	AmqpSession          *amqp.Session
	AmqpReceiver         *amqp.Receiver
}

// Todo: Reconsider approach to error handling in this code.
// Move to returning err and panicing in the caller if listener creation fails.

// NewListener initilises a servicebus lister from configuration
func NewListener(ctx context.Context, config types.Configuration) *Listener {
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

	listener.subscriptionsClient = subsClient
	listener.topicsClient = topicsClient
	listener.namespaceClient = namespaceClient

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
	topic, err := topicsClient.Get(ctx, config.ResourceGroup, config.ServiceBusNamespace, config.SubscribesToEvent)
	if err != nil && topic.Response.StatusCode == http.StatusNotFound {
		log.WithField("config", types.RedactConfigSecrets(config)).Debugf("topic %v doesn't exist.. creating", config.SubscribesToEvent)
		topic, err = topicsClient.CreateOrUpdate(ctx, config.ResourceGroup, config.ServiceBusNamespace, config.SubscribesToEvent, servicebus.SBTopic{})
		if err != nil {
			log.WithField("config", types.RedactConfigSecrets(config)).Panicf("Failed creating topic: %v", err)
		}
	} else if err != nil {
		log.WithField("config", types.RedactConfigSecrets(config)).Panicf("Failed getting topic: %v", err)
	}
	listener.TopicName = strings.ToLower(*topic.Name)

	// Check subscription to listen on. Create if missing
	subName := getSubscriptionName(config.SubscribesToEvent, config.ModuleName)
	sub, err := subsClient.Get(
		ctx,
		config.ResourceGroup,
		config.ServiceBusNamespace,
		config.SubscribesToEvent,
		subName,
	)

	if err != nil && sub.Response.StatusCode == http.StatusNotFound {
		log.WithField("config", types.RedactConfigSecrets(config)).Debugf("subscription %v doesn't exist.. creating", subName)
		sub, err = subsClient.CreateOrUpdate(
			ctx,
			config.ResourceGroup,
			config.ServiceBusNamespace,
			config.SubscribesToEvent,
			subName,
			servicebus.SBSubscription{},
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
