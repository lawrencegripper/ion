package module

import (
	"context"
	"fmt"
	"github.com/lawrencegripper/ion/internal/pkg/management/module"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

type createOptions struct {
	provider           string
	name               string
	eventSubscriptions string
	eventPublications  string
	instanceCount      int32
	retryCount         int32
	configMapFilepath  string
	moduleImage        string
	moduleImageTag     string
	handlerImage       string
	handlerImageTag    string
}

var createOpts createOptions

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a module in ion",
	RunE:  Create,
}

// Create a new ion module
func Create(cmd *cobra.Command, args []string) error {

	var configMap map[string]string
	if createOpts.configMapFilepath != "" {
		var err error
		configMap, err = godotenv.Read(createOpts.configMapFilepath)
		if err != nil {
			return fmt.Errorf("failed to read config map file %s: %+v", createOpts.configMapFilepath, err)
		}
	}

	createRequest := &module.ModuleCreateRequest{
		Modulename:         createOpts.name,
		Eventsubscriptions: createOpts.eventSubscriptions,
		Eventpublications:  createOpts.eventPublications,
		Moduleimage:        createOpts.moduleImage,
		Moduleimagetag:     createOpts.moduleImageTag,
		Handlerimage:       createOpts.handlerImage,
		Handlerimagetag:    createOpts.handlerImageTag,
		Instancecount:      createOpts.instanceCount,
		Retrycount:         createOpts.retryCount,
		Provider:           createOpts.provider,
		Configmap:          configMap,
	}

	fmt.Println("creating module")
	createResponse, err := Client.Create(context.Background(), createRequest)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("failed to create module: %+v", err))
	}
	fmt.Printf("created module %s\n", createResponse.Name)
	return nil
}

func init() {

	// Local flags for the create command
	createCmd.Flags().StringVarP(&createOpts.name, "name", "n", "", "the module name")
	createCmd.Flags().StringVarP(&createOpts.eventSubscriptions, "event-subscriptions", "i", "", "events to which the module subscribes")
	createCmd.Flags().StringVarP(&createOpts.eventPublications, "event-publications", "o", "", "the events the module can publish")
	createCmd.Flags().StringVarP(&createOpts.moduleImage, "module-image", "m", "", "the docker image for your module")
	createCmd.Flags().StringVar(&createOpts.moduleImageTag, "module-image-tag", "latest", "the image tag to use for your module")
	createCmd.Flags().StringVar(&createOpts.configMapFilepath, "config-map-file", "", "a .env file defining environment variables required by the module")
	createCmd.Flags().StringVar(&createOpts.handlerImage, "handler-image", "dotjson/ion-handler", "the docker image for your module")
	createCmd.Flags().StringVar(&createOpts.handlerImageTag, "handler-tag", "latest", "the image tag to use for the handler docker image")
	createCmd.Flags().StringVarP(&createOpts.provider, "provider", "p", "kubernetes", "provider for modules compute resouces (Kubernetes, AzureBatch)")
	createCmd.Flags().Int32Var(&createOpts.instanceCount, "instance-count", 1, "the number of dispatcher instance to create")
	createCmd.Flags().Int32Var(&createOpts.retryCount, "retry-count", 1, "the number of dispatcher instance to create")

	// Mark requried flags
	createCmd.MarkFlagRequired("name")                //nolint: errcheck
	createCmd.MarkFlagRequired("event-subscriptions") //nolint: errcheck
	createCmd.MarkFlagRequired("event-publications")  //nolint: errcheck
	createCmd.MarkFlagRequired("module-image")        //nolint: errcheck
}
