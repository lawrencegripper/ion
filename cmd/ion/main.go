package main

import (
	"github.com/lawrencegripper/ion/cmd/ion/event"
	"github.com/lawrencegripper/ion/cmd/ion/module"
	"github.com/lawrencegripper/ion/cmd/ion/root"
)

func main() {
	module.Register()
	event.Register()
	root.Execute()
}
