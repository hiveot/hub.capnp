package main

import (
	"github.com/sirupsen/logrus"

	"github.com/wostzone/wost-go/pkg/logging"
	"github.com/wostzone/wost-go/pkg/proc"

	"github.com/wostzone/hub/cmd/launcher/internal"
)

// Binary to launch the hub services with their sidecars
func main() {
	logging.SetLogging("warning", "")
	err := internal.StartHub("", true)
	if err != nil {
		logrus.Fatalf("launcher: Failed starting launcher: %s", err)
	}
	proc.WaitForSignal()
	internal.StopHub()
}
