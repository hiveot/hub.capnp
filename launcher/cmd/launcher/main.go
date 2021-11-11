package main

import (
	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/launcher/internal"
	"github.com/wostzone/hub/lib/client/pkg/proc"
)

// Binary to launch the hub services
func main() {
	err := internal.StartHub("", true)
	if err != nil {
		logrus.Fatalf("launcher: Failed starting launcher: %s", err)
	}
	proc.WaitForSignal()
	internal.StopHub()
}
