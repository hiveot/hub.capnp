package main

import (
	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/pkg/hub"
)

func main() {
	err := hub.StartHub("", true)
	if err != nil {
		logrus.Fatalf("main: Failed starting hub: %s", err)
	}
	hub.WaitForSignal()
	hub.StopHub()
}
