package main

import (
	"github.com/lawrencegripper/ion/cmd/ion/dev"
	"github.com/lawrencegripper/ion/cmd/ion/event"
	"github.com/lawrencegripper/ion/cmd/ion/module"
	"github.com/lawrencegripper/ion/cmd/ion/root"
)

func main() {

	// Register commands with root
	module.Register()
	event.Register()
	dev.Register()

	// Execute root
	root.Execute()
}
