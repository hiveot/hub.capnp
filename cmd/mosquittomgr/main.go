// package main for the mosquitto manager
package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hub/core/mosquittomgr"
	"github.com/wostzone/wostlib-go/pkg/hubclient"
	"github.com/wostzone/wostlib-go/pkg/hubconfig"
)

// Main entry to WoST plugin for managing Mosquitto
// This setup the configuration from file and commandline parameters and launches the service
func main() {
	svc := mosquittomgr.NewMosquittoManager()
	hubConfig, err := hubconfig.LoadCommandlineConfig("", mosquittomgr.PluginID, &svc.Config)
	if err != nil {
		logrus.Errorf("Mosquittomgr: Start aborted due to commandline error")
		os.Exit(1)
	}

	err = svc.Start(hubConfig)
	if err != nil {
		logrus.Errorf("Mosquittomgr: Failed to start: %s", err)
		os.Exit(1)
	}
	hubclient.WaitForSignal()
	svc.Stop()
	os.Exit(0)
}
