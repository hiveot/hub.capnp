package main

import (
	"github.com/sirupsen/logrus"
	hub "github.com/wostzone/hub/core/hub"
	"github.com/wostzone/wostlib-go/pkg/hubclient"
)

func main() {
	err := hub.StartHub("", true)
	if err != nil {
		logrus.Fatalf("hub: Failed starting hub: %s", err)
	}
	hubclient.WaitForSignal()
	hub.StopHub()
}
