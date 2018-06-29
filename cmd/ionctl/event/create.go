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

type createOptions struct {
	eventName     string
	eventID       string
	correlationID string
	parentEventID string
	eventType     string
	data          map[string]string
	timeout       int
}

var createOpts createOptions

// createCmd represents the fire command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create an event in the system",
	RunE:  Create,
}

// Create a new ion event
func Create(cmd *cobra.Command, args []string) error {
	sender, err := amqpSession.NewSender(
		amqp.LinkTargetAddress(createOpts.eventType),
	)
	if err != nil {
		return fmt.Errorf("creating sender link: %+v", err)
	}

	var kvps common.KeyValuePairs
	for k, v := range createOpts.data {
		kvp := common.KeyValuePair{
			Key:   k,
			Value: v,
		}
		kvps = kvps.Append(kvp)
	}

	var event = common.Event{
		Context: &common.Context{
			Name:          createOpts.eventName,
			EventID:       createOpts.eventID,
			CorrelationID: createOpts.correlationID,
			ParentEventID: createOpts.parentEventID,
		},
		Type:           createOpts.eventType,
		PreviousStages: []string{},
		Data:           kvps,
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(createOpts.timeout)*time.Second)

	b, err := json.Marshal(&event)
	if err != nil {
		cancel()
		return fmt.Errorf("error encoding event: %+v", err)
	}

	err = sender.Send(ctx, amqp.NewMessage(b))
	if err != nil {
		cancel()
		return fmt.Errorf("error sending message: %+v", err)
	}

	sender.Close(ctx) //nolint: errcheck
	cancel()

	return nil
}

func init() {

	// Local flags for the create command
	createCmd.Flags().StringVar(&createOpts.eventName, "event-name", "test_event", "the name of the event to create (default: test_event)")
	createCmd.Flags().StringVar(&createOpts.eventID, "event-id", "", "ID used to unqiuely identify this event")
	createCmd.Flags().StringVar(&createOpts.eventType, "event-type", "", "the type of event")
	createCmd.Flags().StringVar(&createOpts.parentEventID, "parent-event-id", "", "ID of the parent event")
	createCmd.Flags().StringVar(&createOpts.correlationID, "correlation-id", "", "ID used to unqiuely identify all events in this ion instance")
	createCmd.Flags().IntVar(&createOpts.timeout, "timeout", 30, "timeout in seconds for a connection to the messaging bus")

	// Mark required flags
	createCmd.MarkFlagRequired("event-name")     //nolint: errcheck
	createCmd.MarkFlagRequired("event-id")       //nolint: errcheck
	createCmd.MarkFlagRequired("event-type")     //nolint: errcheck
	createCmd.MarkFlagRequired("correlation-id") //nolint: errcheck
}
