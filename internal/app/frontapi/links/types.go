package links

import (
	"context"

	"github.com/lawrencegripper/ion/internal/pkg/servicebus"
	"github.com/lawrencegripper/ion/internal/pkg/types"
)

var amqpClt *servicebus.AmqpConnection

func InitAmqp(cfg *types.Configuration) {
	amqpClt = servicebus.NewAmqpConnection(context.Background(), cfg)
}

type request struct {
	URL string `json:"url"`
}
