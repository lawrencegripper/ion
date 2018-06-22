package main

import (
	"github.com/lawrencegripper/ion/cmd/ionctl/event"
	"github.com/lawrencegripper/ion/cmd/ionctl/module"
	"github.com/lawrencegripper/ion/cmd/ionctl/root"
)

func main() {
	module.Register()
	event.Register()
	root.Execute()
}
