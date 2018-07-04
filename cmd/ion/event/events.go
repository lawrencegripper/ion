package event

import (
	"fmt"
	"github.com/lawrencegripper/ion/cmd/ion/root"

	"github.com/spf13/cobra"
	"pack.ag/amqp"
)

var amqpSession *amqp.Session
var amqpConnString string

// eventsCmd represents the events command
var eventCmd = &cobra.Command{
	Use:               "event",
	Short:             "execute commands to manage ion events",
	PersistentPreRunE: Setup,
	RunE:              Event,
}

// Setup will run before the event command runs
func Setup(cmd *cobra.Command, args []string) error {

	if cmd.HasSubCommands() {
		return nil
	}

	amqpClient, err := amqp.Dial(amqpConnString)
	if err != nil {
		return fmt.Errorf("error dialing AQMP server: %+v", err)
	}
	amqpSession, err = amqpClient.NewSession()
	if err != nil {
		return fmt.Errorf("error creating AMQP session with server: %+v", err)
	}
	return nil
}

// Event will run when the event subcommand is invoked
func Event(cmd *cobra.Command, args []string) error {
	cmd.Help() // nolint: errcheck
	return nil
}

// Register adds to root command
func Register() {

	// Add event sub commands
	eventCmd.AddCommand(createCmd)
	eventCmd.AddCommand(peekCmd)
	eventCmd.AddCommand(getCmd)

	// Add event command to root
	root.RootCmd.AddCommand(eventCmd)
}

func init() {

	// Local flags to the event command
	eventCmd.PersistentFlags().StringVar(&amqpConnString, "amqp-connection-string", "", "AMQP connection string")

	// Mark requried flags
	eventCmd.MarkPersistentFlagRequired("amqp-connection-string") //nolint: errcheck
}
