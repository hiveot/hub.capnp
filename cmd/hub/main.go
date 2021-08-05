package main

import (
	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/pkg/hub"
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
