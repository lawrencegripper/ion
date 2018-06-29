package main

import (
	"github.com/lawrencegripper/ion/cmd/ionctl/dev"
	"github.com/lawrencegripper/ion/cmd/ionctl/event"
	"github.com/lawrencegripper/ion/cmd/ionctl/module"
	"github.com/lawrencegripper/ion/cmd/ionctl/root"
)

func main() {

	// Register commands with root
	module.Register()
	event.Register()
	dev.Register()

	// Execute root
	root.Execute()
}
