package module

import (
	"github.com/spf13/cobra"
)

var provider string
var name string
var eventSubscriptions string
var eventPublications string
var instanceCount int
var retryCount int
var configMapFilepath string
var moduleImage string
var handlerImage string
var handlerTag string

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a module in ion",
	Run: func(cmd *cobra.Command, args []string) {
		//todo: add logic
	},
}

func init() {
	moduleCmd.AddCommand(createCmd)

	createCmd.Flags().StringVarP(&name, "name", "n", "", "The module name")
	createCmd.MarkFlagRequired("name")
	createCmd.Flags().StringVarP(&eventSubscriptions, "event-subscriptions", "i", "", "Events to which the module subscribes")
	createCmd.MarkFlagRequired("event-subscriptions")
	createCmd.Flags().StringVarP(&eventPublications, "events-publications", "o", "", "The events the module can publish")
	createCmd.MarkFlagRequired("event-publications")

	createCmd.Flags().StringVar(&moduleImage, "module-image", "", "The docker image for your module")
	createCmd.MarkFlagRequired("module-image")
	createCmd.Flags().StringVar(&handlerImage, "handler-image", "", "The docker image for your module")
	createCmd.MarkFlagRequired("handler-image")

	createCmd.Flags().StringVar(&handlerTag, "handler-tag", "latest", "The tag to use for the handler docker image")
	createCmd.Flags().StringVarP(&provider, "provider", "p", "kubernetes", "Provider for modules compute resouces (Kubernetes, AzureBatch)")
	createCmd.Flags().IntVar(&instanceCount, "instance-count", 1, "The number of dispatcher instance to create")
	createCmd.Flags().IntVar(&retryCount, "retry-count", 1, "The number of dispatcher instance to create")
}
