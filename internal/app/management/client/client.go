// nolint: errcheck
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/lawrencegripper/ion/internal/pkg/management/module"
	"google.golang.org/grpc"
)

func main() {
	if err := runClient(); err != nil {
		fmt.Fprintf(os.Stderr, "failed: %v\n", err)
		os.Exit(1)
	}
}

func runClient() error {
	conn, err := grpc.Dial("localhost:9000", grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("failed to dial server: %+v", err)
	}
	cl := module.NewModuleServiceClient(conn)

	// Create new module
	createRequest := &module.ModuleCreateRequest{
		Modulename:         "testmodule",
		Eventsubscriptions: "new_video",
		Eventpublications:  "face_detected",
		Moduleimage:        "dotjson/ion-python-example-module",
		Moduleimagetag:     "latest",
		Handlerimage:       "dotjson/handler",
		Handlerimagetag:    "latest",
		Instancecount:      1,
		Retrycount:         3,
		Provider:           "kubernetes",
		Configmap: map[string]string{
			"HANDLER_BASE_DIR": "/ion/",
		},
	}

	fmt.Println("Creating module")
	createResponse, err := cl.Create(context.Background(), createRequest)
	if err != nil {
		return fmt.Errorf("failed to create module: %+v", err)
	}
	fmt.Printf("Created module %s\n", createResponse.Name)

	time.Sleep(5 * time.Second)

	// List all modules
	listRequest := &module.ModuleListRequest{}

	fmt.Println("Listing all modules")
	listResponse, err := cl.List(context.Background(), listRequest)
	if err != nil {
		return fmt.Errorf("failed to list module: %+v", err)
	}
	for _, name := range listResponse.Names {
		fmt.Printf("%s\n", name)
	}

	time.Sleep(5 * time.Second)

	// Get module
	moduleIsAvailable := false
	for !moduleIsAvailable {
		fmt.Printf("Getting module %s\n", createResponse.Name)
		getRequest := &module.ModuleGetRequest{
			Name: createResponse.Name,
		}
		getResponse, err := cl.Get(context.Background(), getRequest)
		if err != nil {
			return fmt.Errorf("Failed to get module %s with error %+v", createResponse.Name, err)
		}
		fmt.Printf("Got module %s, status: %s, Message: %s\n", getResponse.Name, getResponse.Status, getResponse.StatusMessage)
		if getResponse.Status == "Available" {
			moduleIsAvailable = true
			break
		}
		time.Sleep(1 * time.Second)
	}

	// Delete module
	fmt.Printf("Deleting module %s\n", createResponse.Name)
	deleteRequest := &module.ModuleDeleteRequest{
		Name: createResponse.Name,
	}
	deleteResponse, err := cl.Delete(context.Background(), deleteRequest)
	fmt.Printf("Deleted module %s\n", deleteResponse.Name)

	return nil
}
