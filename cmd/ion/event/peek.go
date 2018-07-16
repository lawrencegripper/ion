package event

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lawrencegripper/ion/internal/pkg/common"
	"github.com/spf13/cobra"
	"pack.ag/amqp"
	"time"
)

type peekOptions struct {
	subName   string
	eventType string
	timeout   int
}

var peekOpts peekOptions

// peekCmd represents the fire command
var peekCmd = &cobra.Command{
	Use:   "peek",
	Short: "peek an event from the messaging queue",
	RunE:  Peek,
}

// Peek takes the latest event of the queue but does not remove it
func Peek(cmd *cobra.Command, args []string) error {

	subPath := fmt.Sprintf("/%s/subscriptions/%s", peekOpts.eventType, peekOpts.subName)

	receiver, err := amqpSession.NewReceiver(
		amqp.LinkSourceAddress(subPath),
	)
	if err != nil {
		return fmt.Errorf("error creating receiver link: %+v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(peekOpts.timeout)*time.Second)

	defer func() {
		cancel()
		ctx, cancel := context.WithTimeout(ctx, time.Duration(peekOpts.timeout)*time.Second)
		receiver.Close(ctx) //nolint: errcheck
		cancel()
	}()

	// Receive next event
	msg, err := receiver.Receive(ctx)
	if err != nil {
		return fmt.Errorf("error reading event: %+v", err)
	}

	if msg == nil {
		fmt.Println("no events available")
		return nil
	}

	// Release event
	msg.Release()

	var event common.Event
	b := msg.GetData()
	if err := json.Unmarshal(b, &event); err != nil {
		return fmt.Errorf("error decoding event: %+v", err)
	}

	fmt.Println(string(msg.GetData()))
	return nil
}

func init() {

	// Local flags for the create command
	peekCmd.Flags().StringVar(&peekOpts.subName, "sub-name", "", "the name of AQMP subscription to read from")
	peekCmd.Flags().StringVar(&peekOpts.eventType, "event-type", "", "the event type name")
	peekCmd.Flags().IntVar(&peekOpts.timeout, "timeout", 15, "timeout in seconds for a connection to the messaging bus")

	// Mark required flags
	peekCmd.MarkFlagRequired("sub-name")   //nolint: errcheck
	peekCmd.MarkFlagRequired("event-type") //nolint: errcheck
}
