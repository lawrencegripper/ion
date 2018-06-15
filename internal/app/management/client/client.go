package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/lawrencegripper/ion/internal/app/management/module"
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
		Name:                      "test",
		Eventsubscriptions:        "new_video",
		Eventpublications:         "face_detected",
		Moduleimage:               "dotjson/ion-python-example-module",
		Moduleimagetag:            "latest",
		Handlerimage:              "dotjson/handler",
		Handlerimagetag:           "latest",
		Instancecount:             1,
		Retrycount:                3,
		Containerregistryurl:      "https://hub.docker.com",
		Containerregistryusername: "",
		Containerregistryemail:    "",
		Containerregistrypassword: "",
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

	listRequest := &module.ModuleListRequest{}

	fmt.Println("Listing all modules")
	listResponse, err := cl.List(context.Background(), listRequest)
	if err != nil {
		return fmt.Errorf("failed to list module: %+v", err)
	}
	for _, moduleName := range listResponse.Names {
		fmt.Printf("%s\n", moduleName)
	}

	time.Sleep(5 * time.Second)

	fmt.Printf("Deleting module %s\n", createRequest.Name)
	deleteRequest := &module.ModuleDeleteRequest{
		Name: createResponse.Name,
	}
	deleteResponse, err := cl.Delete(context.Background(), deleteRequest)
	fmt.Printf("Deleted module %s\n", deleteResponse.Name)

	return nil
}
