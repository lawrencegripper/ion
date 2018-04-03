package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/servicebus/mgmt/2017-04-01/servicebus"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/lawrencegripper/ion/dispatcher/types"
	"pack.ag/amqp"
)

func main() {

	fmt.Println("Initializing subscriber")

	// Parse configuration
	subscriptionID := flag.String("azuresubid", "", "Azure Subscription ID")
	tenantID := flag.String("azuretenantid", "", "Azure Tenant ID")
	clientID := flag.String("azureclientid", "", "Azure Client ID")
	resourceGroup := flag.String("azureresourcegroup", "", "Azure Resource Group")
	clientSecret := flag.String("azureclientsecret", "", "Azure Client secret")
	primaryKey := flag.String("amqpprimarykey", "", "AMQP Primary Key")
	namespace := flag.String("amqpnamespace", "", "AMQP namespace")
	topicName := flag.String("amqptopicname", "", "AMQP topic name")
	moduleName := flag.String("modulename", "", "Module name")
	retryCount := 5
	flag.Parse()

	if subscriptionID == nil || tenantID == nil || clientID == nil || resourceGroup == nil || clientSecret == nil ||
		primaryKey == nil || namespace == nil || topicName == nil || moduleName == nil {
		panic("Missing configuration!")
	}

	ctx := context.Background()

	// Get Service Bus topic client
	auth := getAuthorizer(*tenantID, *clientID, *clientSecret)
	topicsClient := servicebus.NewTopicsClient(*subscriptionID)
	topicsClient.Authorizer = auth

	// Get Service Bus subscriptions client
	subsClient := servicebus.NewSubscriptionsClient(*subscriptionID)
	subsClient.Authorizer = auth

	// Get Service Bus subscription
	subName := getSubscriptionName(*topicName, *moduleName)
	sub, err := subsClient.Get(
		ctx,
		*resourceGroup,
		*namespace,
		*topicName,
		subName,
	)

	if err != nil && sub.Response.StatusCode == http.StatusNotFound {
		// Create subscription if it doesn't exist
		fmt.Printf("subscription %v doesn't exist.. creating", subName)
		subDef := servicebus.SBSubscription{
			SBSubscriptionProperties: &servicebus.SBSubscriptionProperties{
				MaxDeliveryCount: to.Int32Ptr(int32(retryCount)),
			},
		}
		sub, err = subsClient.CreateOrUpdate(
			ctx,
			*resourceGroup,
			*namespace,
			*topicName,
			subName,
			subDef,
		)
		if err != nil {
			panic(fmt.Sprintf("Failed creating subscription: %v", err))
		}
	} else if err != nil {
		panic(fmt.Sprintf("Failed getting subscription: %v", err))
	}

	// Create AMQP session
	amqpConnectionString := getAmqpConnectionString("RootManageSharedAccessKey", *primaryKey, *namespace)
	amqpSession := createAmqpSession(amqpConnectionString)

	// Get AMQP receiver
	subscriptionAmqpPath := getSubscriptionAmqpPath(*topicName, *moduleName)
	amqpReceiver := createAmqpListener(amqpSession, subscriptionAmqpPath)

	// Listen to messages in a infinite loop
	fmt.Printf("Subscription '%s' initialized, starting to listen...\n", subName)
	for {
		message, err := amqpReceiver.Receive(ctx)
		if err != nil {
			panic(fmt.Sprintf("Error received dequeuing message"))
		}
		if message == nil {
			panic(fmt.Sprintf("Error received dequeuing message - nil message"))
		}
		data := message.GetData()

		var event Event
		err = json.Unmarshal(data, &event)
		if err != nil {
			fmt.Printf("Error unmarshalling event: '%+v'\n", err)
			message.Reject()
		} else {
			event.ID = fmt.Sprintf("%v", message.Properties.MessageID)
			fmt.Printf("Event received: '%+v'\n", event)
			message.Accept()
		}
	}
}

//Event is the expected event structure of messages on the topic
type Event struct {
	ID             string            `json:"id"`
	Type           string            `json:"type"`
	PreviousStages []string          `json:"previousStages"`
	ParentEventID  string            `json:"parentId"`
	CorrelationID  string            `json:"correlationId"`
	Data           map[string]string `json:"data"`
}

func createAmqpSession(amqpConnectionString string) *amqp.Session {
	client, err := amqp.Dial(amqpConnectionString)
	if err != nil {
		panic(fmt.Sprintf("Dialing AMQP server: '%+v'", err))
	}
	session, err := client.NewSession()
	if err != nil {
		panic("Creating session failed")
	}

	return session
}

func createAmqpListener(session *amqp.Session, subscriptionAmqpPath string) *amqp.Receiver {
	receiver, err := session.NewReceiver(
		amqp.LinkSourceAddress(subscriptionAmqpPath),
		// amqp.LinkCredit(10), // Todo: Add config value to define how many inflight tasks the dispatcher can handle
	)
	if err != nil {
		panic(fmt.Sprintf("Creating receiver: '+%v'", err))
	}

	return receiver
}

func getOrCreateTopic(ctx context.Context, topicsClient servicebus.TopicsClient, config *types.Configuration, topicName string) servicebus.SBTopic {
	topic, err := topicsClient.Get(ctx, config.ResourceGroup, config.ServiceBusNamespace, topicName)
	if err != nil && topic.Response.StatusCode == http.StatusNotFound {
		topic, err = topicsClient.CreateOrUpdate(ctx, config.ResourceGroup, config.ServiceBusNamespace, topicName, servicebus.SBTopic{})
		if err != nil {
			panic(fmt.Sprintf("Failed creating topic: %v", err))
		}
	} else if err != nil {
		panic(fmt.Sprintf("Failed getting topic: %v", err))
	}

	return topic
}

func getAmqpConnectionString(keyName, keyValue, namespace string) string {
	encodedKey := url.QueryEscape(keyValue)
	return fmt.Sprintf("amqps://%s:%s@%s.servicebus.windows.net", keyName, encodedKey, namespace)
}

func getSubscriptionName(eventName, moduleName string) string {
	return strings.ToLower(eventName) + "_" + strings.ToLower(moduleName)
}

func getSubscriptionAmqpPath(eventName, moduleName string) string {
	return "/" + strings.ToLower(eventName) + "/subscriptions/" + getSubscriptionName(eventName, moduleName)
}

func newServicePrincipalTokenFromCredentials(tenantID, clientID, clientSecret, scope string) (*adal.ServicePrincipalToken, error) {
	oauthConfig, err := adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, tenantID)
	if err != nil {
		panic(err)
	}
	return adal.NewServicePrincipalToken(*oauthConfig, clientID, clientSecret, scope)
}

func getAuthorizer(tenantID, clientID, clientSecret string) autorest.Authorizer {
	spt, err := newServicePrincipalTokenFromCredentials(tenantID, clientID, clientSecret, azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		panic(fmt.Sprintf("Failed to create authorizer: %v", err))
	}
	auth := autorest.NewBearerAuthorizer(spt)
	return auth
}
