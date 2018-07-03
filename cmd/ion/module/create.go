package module

import (
	"context"
	"fmt"
	"github.com/lawrencegripper/ion/cmd/ion/root"
	"github.com/lawrencegripper/ion/internal/pkg/management/module"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var provider string
var name string
var eventSubscriptions string
var eventPublications string
var instanceCount int32
var retryCount int32
var configMapFilepath string
var moduleImage string
var handlerImage string
var handlerTag string

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a module in ion",
	Run: func(cmd *cobra.Command, args []string) {
		conn, err := grpc.Dial(root.ManagementAPIEndpoint, grpc.WithInsecure())
		if err != nil {
			fmt.Println(fmt.Sprintf("failed to dial server: %+v", err))
			return
		}
		cl := module.NewModuleServiceClient(conn)

		// Create new module
		createRequest := &module.ModuleCreateRequest{
			Modulename:         name,
			Eventsubscriptions: eventSubscriptions,
			Eventpublications:  eventPublications,
			Moduleimage:        moduleImage,
			Moduleimagetag:     "latest",
			Handlerimage:       handlerImage,
			Handlerimagetag:    "latest",
			Instancecount:      instanceCount,
			Retrycount:         retryCount,
			Provider:           provider,
			Configmap: map[string]string{
				"HANDLER_BASE_DIR": "/ion/",
			},
		}

		fmt.Println("Creating module")
		createResponse, err := cl.Create(context.Background(), createRequest)
		if err != nil {
			fmt.Println(fmt.Sprintf("failed to create module: %+v", err))
			return
		}
		fmt.Printf("Created module %s\n", createResponse.Name)

	},
}

func init() {
	moduleCmd.AddCommand(createCmd)

	createCmd.Flags().StringVarP(&name, "name", "n", "", "The module name")
	createCmd.MarkFlagRequired("name") //nolint: errcheck
	createCmd.Flags().StringVarP(&eventSubscriptions, "event-subscriptions", "i", "", "Events to which the module subscribes")
	createCmd.MarkFlagRequired("event-subscriptions") //nolint: errcheck
	createCmd.Flags().StringVarP(&eventPublications, "events-publications", "o", "", "The events the module can publish")
	createCmd.MarkFlagRequired("event-publications") //nolint: errcheck
	createCmd.Flags().StringVarP(&moduleImage, "module-image", "m", "", "The docker image for your module")
	createCmd.MarkFlagRequired("module-image") //nolint: errcheck

	createCmd.Flags().StringVar(&configMapFilepath, "config-map-file", "", "A .env file defining environment variables required by the module")

	createCmd.Flags().StringVar(&handlerImage, "handler-image", "dotjson/ion-handler", "The docker image for your module")
	createCmd.Flags().StringVar(&handlerTag, "handler-tag", "latest", "The tag to use for the handler docker image")
	createCmd.Flags().StringVarP(&provider, "provider", "p", "kubernetes", "Provider for modules compute resouces (Kubernetes, AzureBatch)")
	createCmd.Flags().Int32Var(&instanceCount, "instance-count", 1, "The number of dispatcher instance to create")
	createCmd.Flags().Int32Var(&retryCount, "retry-count", 1, "The number of dispatcher instance to create")
}
