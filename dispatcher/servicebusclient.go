package main

import (
	"log"

	"github.com/Azure/Azure-sdk-for-go/services/servicebus/mgmt/2017-04-01/servicebus"
)

func startListening() {
	client := servicebus.New("thing")
	log.Println(client)
}
