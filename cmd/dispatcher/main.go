package main

import "os"

func main() {
	dispatcherCmd := NewDispatcherCommand()

	if err := dispatcherCmd.Execute(); err != nil {
		//TODO Should I activate the silentError of cobra and print error with logrus?
		//     Or is it fine until it's PreRun errors and that why I shouldn't use RunE?
		//log.Fatalf("ion-dispatcher error: %v\n", err)
		os.Exit(1)
	}
}
