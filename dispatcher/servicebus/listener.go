package servicebus

import (
	log "github.com/sirupsen/logrus"

	"github.com/Azure/Azure-sdk-for-go/services/servicebus/mgmt/2017-04-01/servicebus"
	"github.com/lawrencegripper/mlops/dispatcher/types"
)

// Listener provides a connection to service bus and methods for creating required subscriptions and topics
type Listener struct {
	namespaceClient     servicebus.NamespacesClient
	subscriptionsClient servicebus.SubscriptionsClient
	topicsClient        servicebus.TopicsClient
}

// NewListener initilises a servicebus lister from configuration
func NewListener(config types.Configuration) Listener {
	listener := Listener{}
	auth := getAuthorizer(config)
	subsClient := servicebus.NewSubscriptionsClient(config.SubscriptionID)
	subsClient.Authorizer = auth
	topicsClient := servicebus.NewTopicsClient(config.SubscriptionID)
	topicsClient.Authorizer = auth
	namespaceClient := servicebus.NewNamespacesClient(config.SubscriptionID)
	namespaceClient.Authorizer = auth

	listener.subscriptionsClient = subsClient
	listener.topicsClient = topicsClient
	listener.namespaceClient = namespaceClient

	// Check subscription to listen on. Create a topic and sub if missing

	return listener
}

// Start starts listening to the bus for new messages
func Start() {
	client := servicebus.New("thing")
	log.Println(client)
}
