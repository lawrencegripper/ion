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

type getOptions struct {
	subName   string
	eventType string
	timeout   int
}

var getOpts getOptions

// getCmd represents the fire command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get an event from the messaging queue (dequeue)",
	RunE:  Get,
}

// Get an event from the ion eventing system (dequeues the message)
func Get(cmd *cobra.Command, args []string) error {

	subPath := fmt.Sprintf("/%s/subscriptions/%s", getOpts.eventType, getOpts.subName)

	receiver, err := amqpSession.NewReceiver(
		amqp.LinkSourceAddress(subPath),
	)
	if err != nil {
		return fmt.Errorf("error creating receiver link: %+v", err)
	}

	ctx := context.Background()

	defer func() {
		ctx, cancel := context.WithTimeout(ctx, time.Duration(getOpts.timeout)*time.Second)
		receiver.Close(ctx) //nolint: errcheck
		cancel()
	}()

	// Receive next event
	msg, err := receiver.Receive(ctx)
	if err != nil {
		return fmt.Errorf("error reading event: %+v", err)
	}

	if msg == nil {
		return fmt.Errorf("nil message returned by receiver")
	}

	var event common.Event
	b := msg.GetData()
	if err := json.Unmarshal(b, &event); err != nil {
		return fmt.Errorf("error decoding event: %+v", err)
	}

	// Dequeue event
	err = msg.Accept()
	if err != nil {
		return fmt.Errorf("error accepting event: %+v", err)
	}

	fmt.Println(string(msg.GetData()))
	return nil
}

func init() {

	// Local flags for the create command
	getCmd.Flags().StringVar(&getOpts.subName, "sub-name", "", "the name of AQMP subscription to read from")
	getCmd.Flags().StringVar(&getOpts.eventType, "event-type", "", "the event type name")
	getCmd.Flags().IntVar(&getOpts.timeout, "timeout", 30, "timeout in seconds for a connection to the messaging bus")

	// Mark required flags
	getCmd.MarkFlagRequired("sub-name")   //nolint: errcheck
	getCmd.MarkFlagRequired("event-type") //nolint: errcheck
}
