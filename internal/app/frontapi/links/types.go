package links

import (
	"context"

	"github.com/lawrencegripper/ion/internal/pkg/servicebus"
	"pack.ag/amqp"

	"github.com/lawrencegripper/ion/internal/pkg/types"

	"github.com/lawrencegripper/ion/internal/app/handler/dataplane/documentstorage/mongodb"
)

var amqpClt *servicebus.AmqpConnection
var amqpSender *amqp.Sender
var eventType string

// InitAmqp initialize the amqp client to fire events
func InitAmqp(cfg *types.Configuration, eventToSend string) {
	amqpClt = servicebus.NewAmqpConnection(context.Background(), cfg)
	eventType = eventToSend
	newSender, err := amqpClt.CreateAmqpSender(eventToSend)
	if err != nil {
		panic(err)
	}
	amqpSender = newSender
}

type request struct {
	URL string `json:"url"`
}

var documentStore *mongodb.MongoDB

// InitMongoDB initialize the mongodb connection for storing event data
func InitMongoDB(cfg *types.Configuration) {
	docStore, err := mongodb.NewMongoDB(&mongodb.Config{
		Enabled:    true,
		Name:       cfg.Handler.MongoDBDocumentStorageProvider.Name,
		Collection: cfg.Handler.MongoDBDocumentStorageProvider.Collection,
		Password:   cfg.Handler.MongoDBDocumentStorageProvider.Password,
		Port:       cfg.Handler.MongoDBDocumentStorageProvider.Port,
	})

	if err != nil {
		panic("Can't connect to mongodb")
	}

	documentStore = docStore
}
