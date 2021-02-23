package main

import (
	"github.com/sirupsen/logrus"
	"github.com/wostzone/gateway/pkg/lib"
	"github.com/wostzone/gateway/plugins/smbus/internal"
)

// parse commandline, load configuration and start the plugin
func main() {
	srv, err := internal.StartSmbus("")
	if err != nil {
		logrus.Errorf("smbus: Failed to start")
		srv.Stop()
	}
	lib.WaitForSignal()
}
