package main

import (
	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/pkg/hub"
	"github.com/wostzone/hubclient-go/pkg/proc"
)

func main() {
	err := hub.StartHub("", true)
	if err != nil {
		logrus.Fatalf("hub: Failed starting hub: %s", err)
	}
	proc.WaitForSignal()
	hub.StopHub()
}
