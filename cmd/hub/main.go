package main

import (
	"github.com/sirupsen/logrus"
	hub "github.com/wostzone/hub/internal"
	"github.com/wostzone/hubapi/pkg/plugin"
)

func main() {
	err := hub.StartHub("", true)
	if err != nil {
		logrus.Fatalf("main: Failed starting hub: %s", err)
	}
	plugin.WaitForSignal()
	hub.StopHub()
}
