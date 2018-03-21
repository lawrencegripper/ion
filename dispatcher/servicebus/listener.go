package servicebus

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/azure/azure-sdk-for-go/services/resources/mgmt/2017-05-10/resources"
	"github.com/azure/azure-sdk-for-go/services/servicebus/mgmt/2017-04-01/servicebus"
	"github.com/lawrencegripper/mlops/dispatcher/types"
	"github.com/streadway/amqp"
)

// Listener provides a connection to service bus and methods for creating required subscriptions and topics
type Listener struct {
	namespaceClient      servicebus.NamespacesClient
	subscriptionsClient  servicebus.SubscriptionsClient
	topicsClient         servicebus.TopicsClient
	Endpoint             string
	SubscriptionName     string
	TopicName            string
	AccessKeys           servicebus.AccessKeys
	AMQPConnectionString string
}

// NewListener initilises a servicebus lister from configuration
func NewListener(ctx context.Context, config types.Configuration) Listener {
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

	keys, err := namespaceClient.ListKeys(ctx, config.ResourceGroup, config.ServiceBusNamespace, "RootManagerSharedAccessKey")
	if err != nil {
		log.WithFields(log.Fields{
			"config":   types.RedactConfigSecrets(config),
			"response": keys,
		}).WithError(err).Panicf("Failed getting servicebus namespace")
	}
	listener.AccessKeys = keys
	listener.AMQPConnectionString = fmt.Sprintf("amqps://%s:%s@%s.servicebus.windows.net", config.ServiceBusNamespace, keys.KeyName, keys.PrimaryConnectionString)

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
	listener.TopicName = *topic.Name

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

	return listener
}

// Start starts listening to the bus for new messages
func Start() {
	client := servicebus.New("thing")
	log.Println(client)
}

func getSubscriptionName(eventName, moduleName string) string {
	return strings.Join([]string{eventName, moduleName}, "_")
}

func addAmqpSender(listener *Listener) <-chan amqp.Delivery {
	// Native AMQP Library
	// Create client
	connection, err := amqp.Dial(listener.AMQPConnectionString)
	if err != nil {
		log.Fatal("Dialing AMQP server:", err)
	}
	defer connection.Close()
	go func() {
		log.Printf("closing: %s", <-connection.NotifyClose(make(chan *amqp.Error)))
	}()

	channel, err := connection.Channel()
	if err != nil {
		log.WithError(err).Panicln("Failed creating amqp channel")
	}

	err = channel.ExchangeDeclare(
		listener.TopicName, // name of the exchange
		"topic",            // type
		true,               // durable
		false,              // delete when complete
		false,              // internal
		false,              // noWait
		nil,                // arguments
	)
	if err != nil {
		log.WithError(err).Panicf("Exchange declaration failed")
	}

	err = channel.QueueBind(
		listener.SubscriptionName, // name of the queue
		"#",                // route all messages
		listener.TopicName, // sourceExchange
		false,              // noWait
		nil,                // arguments
	)
	if err != nil {
		log.WithError(err).Panicf("Queue binding failed")
	}

	deliveries, err := channel.Consume(
		listener.SubscriptionName, // name
		listener.SubscriptionName, // consumerTag,
		false, // noAck
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // arguments
	)
	if err != nil {
		log.WithError(err).Panicf("Queue Consume failed")
	}

	return deliveries
}
