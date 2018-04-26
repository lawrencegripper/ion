package main

import (
	log "github.com/sirupsen/logrus"
)

func main() {
	dispatcherCmd := NewDispatcherCommand()

	if err := dispatcherCmd.Execute(); err != nil {
		log.Fatalf("ion-dispatcher error: %v\n", err)
	}
}
