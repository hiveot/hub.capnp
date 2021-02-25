package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/gateway/pkg/lib"
	"github.com/wostzone/gateway/plugins/smbus/internal"
)

// parse commandline, load configuration and start the plugin
func main() {
	srv, err := internal.StartSmbus("")
	if err != nil {
		logrus.Errorf("smbus: Failed to start")
		os.Exit(1)
	}
	lib.WaitForSignal()
	srv.Stop()
	os.Exit(0)
}
